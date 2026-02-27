package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	"clawclack/bot/pkg/agent"
	"clawclack/bot/pkg/handlers"
	"clawclack/bot/pkg/shkeeper"
)

type Bot struct {
	Client    *mautrix.Client
	Config    *Config
	SHKeeper  *shkeeper.Client
	Agent     *agent.Agent
	Handlers  *handlers.Registry
}

type Config struct {
	Matrix struct {
		Homeserver string `mapstructure:"homeserver"`
		UserID     string `mapstructure:"user_id"`
		Token      string `mapstructure:"access_token"`
		DeviceID   string `mapstructure:"device_id"`
	}
	SHKeeper struct {
		URL    string `mapstructure:"url"`
		APIKey string `mapstructure:"api_key"`
	}
	Agent struct {
		SpendingLimitUSD float64 `mapstructure:"spending_limit_usd"`
		DailyBudgetUSD   float64 `mapstructure:"daily_budget_usd"`
		OpenAIKey        string  `mapstructure:"openai_key"`
	}
	LogLevel string `mapstructure:"log_level"`
}

func main() {
	log.Info("🤖 Starting ClawClack Agent...")

	config := loadConfig()
	setupLogging(config.LogLevel)

	bot, err := NewBot(config)
	if err != nil {
		log.Fatal("Failed to create bot", "error", err)
	}

	if err := bot.Start(); err != nil {
		log.Fatal("Failed to start bot", "error", err)
	}

	// Wait for interrupt
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	log.Info("👋 Shutting down...")
	bot.Stop()
}

func NewBot(config *Config) (*Bot, error) {
	// Create Matrix client
	client, err := mautrix.NewClient(config.Matrix.Homeserver, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to create Matrix client: %w", err)
	}

	client.SetAccessToken(config.Matrix.Token)
	client.DeviceID = id.DeviceID(config.Matrix.DeviceID)

	// Create SHKeeper client
	skClient := shkeeper.New(config.SHKeeper.URL, config.SHKeeper.APIKey)

	// Create AI agent
	aiAgent := agent.New(agent.Config{
		SpendingLimitUSD: config.Agent.SpendingLimitUSD,
		DailyBudgetUSD:   config.Agent.DailyBudgetUSD,
		OpenAIKey:        config.Agent.OpenAIKey,
	})

	bot := &Bot{
		Client:   client,
		Config:   config,
		SHKeeper: skClient,
		Agent:    aiAgent,
	}

	// Register handlers
	bot.Handlers = handlers.NewRegistry()
	bot.registerHandlers()

	return bot, nil
}

func (b *Bot) Start() error {
	log.Info("🚀 Connecting to Matrix...", "homeserver", b.Config.Matrix.Homeserver)

	// Sync filter to only get messages we care about
	filter := &mautrix.Filter{
		Room: struct {
			Rooms     []id.RoomID    `json:"rooms,omitempty"`
			NotRooms  []id.RoomID    `json:"not_rooms,omitempty"`
			Ephemeral mautrix.FilterPart `json:"ephemeral,omitempty"`
			State     mautrix.FilterPart `json:"state,omitempty"`
			Timeline  mautrix.FilterPart `json:"timeline,omitempty"`
			AccountData mautrix.FilterPart `json:"account_data,omitempty"`
		}{
			Timeline: mautrix.FilterPart{
				Types: []event.Type{event.EventMessage},
			},
		},
	}

	b.Client.Syncer = mautrix.NewDefaultSyncer()
	b.Client.Store = &mautrix.MemoryStore{}

	// Set up event handlers
	b.Client.Syncer.(*mautrix.DefaultSyncer).OnEventType(event.EventMessage, b.handleMessage)
	b.Client.Syncer.(*mautrix.DefaultSyncer).OnEventType(event.StateMember, b.handleMembership)

	// Start syncing
	go func() {
		for {
			err := b.Client.Sync()
			if err != nil {
				log.Error("Sync error", "error", err)
				time.Sleep(5 * time.Second)
			}
		}
	}()

	// Set display name
	_, _ = b.Client.SetDisplayName(context.Background(), "ClawClack Agent 🤖")

	log.Info("✅ Bot is running!")
	return nil
}

func (b *Bot) Stop() {
	b.Client.StopSync()
}

func (b *Bot) handleMessage(source mautrix.EventSource, evt *event.Event) {
	// Ignore our own messages
	if evt.Sender.String() == b.Config.Matrix.UserID {
		return
	}

	// Only process text messages
	if evt.Content.AsMessage().MsgType != event.MsgText {
		return
	}

	content := strings.TrimSpace(evt.Content.AsMessage().Body)
	roomID := evt.RoomID
	sender := evt.Sender

	log.Info("📩 Received message", "room", roomID, "sender", sender, "content", content)

	// Route to appropriate handler
	ctx := &handlers.Context{
		Client:   b.Client,
		RoomID:   roomID,
		Sender:   sender,
		Message:  content,
		SHKeeper: b.SHKeeper,
		Agent:    b.Agent,
	}

	if handler := b.Handlers.Find(content); handler != nil {
		go handler.Handle(ctx)
	}
}

func (b *Bot) handleMembership(source mautrix.EventSource, evt *event.Event) {
	if evt.GetStateKey() == b.Config.Matrix.UserID.String() {
		if evt.Content.AsMember().Membership == event.MembershipInvite {
			// Auto-join invited rooms
			log.Info("📨 Auto-joining room", "room", evt.RoomID)
			_, err := b.Client.JoinRoom(context.Background(), evt.RoomID.String(), "", nil)
			if err != nil {
				log.Error("Failed to join room", "error", err)
			} else {
				// Send welcome message
				b.sendWelcome(evt.RoomID)
			}
		}
	}
}

func (b *Bot) sendWelcome(roomID id.RoomID) {
	msg := `👋 Hello! I'm **ClawClack Agent**.

I offer AI-powered services and can help your group:

**Free commands:**
• !help - Show all commands
• !balance - Check my treasury and spending
• !services - What I can do
• !price <crypto> - Get current crypto price

**Paid services:**
• !alert <crypto> <price> - Price alerts ($0.10)
• !summarize <url> - Article summary ($0.50)
• !image <prompt> - AI image generation ($0.75)
• !code <description> - Code generation ($0.50)

Type !help for more details.`

	_, _ = b.Client.SendText(context.Background(), roomID, msg)
}

func (b *Bot) registerHandlers() {
	b.Handlers.Register("!help", &handlers.HelpHandler{})
	b.Handlers.Register("!balance", &handlers.BalanceHandler{})
	b.Handlers.Register("!services", &handlers.ServicesHandler{})
	b.Handlers.Register("!price", &handlers.PriceHandler{})
	b.Handlers.Register("!alert", &handlers.AlertHandler{})
	b.Handlers.Register("!summarize", &handlers.SummarizeHandler{})
	b.Handlers.Register("!image", &handlers.ImageHandler{})
	b.Handlers.Register("!code", &handlers.CodeHandler{})
	b.Handlers.Register("!propose", &handlers.ProposeHandler{})
	b.Handlers.Register("!pay", &handlers.PaymentHandler{})
	b.Handlers.Register("!status", &handlers.StatusHandler{})
}

func loadConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/clawclack/")
	viper.AddConfigPath("$HOME/.clawclack")

	// Environment variables
	viper.SetEnvPrefix("CLAWCLACK")
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("log_level", "info")
	viper.SetDefault("matrix.homeserver", "https://matrix.org")
	viper.SetDefault("agent.spending_limit_usd", 1.0)
	viper.SetDefault("agent.daily_budget_usd", 5.0)

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		log.Warn("No config file found, using environment variables")
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal("Failed to unmarshal config", "error", err)
	}

	return &config
}

func setupLogging(level string) {
	switch level {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}
