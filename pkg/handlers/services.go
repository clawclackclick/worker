package handlers

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
)

// AlertHandler - Price alerts ($0.10)
type AlertHandler struct{}

func (h *AlertHandler) Handle(ctx *Context) error {
	// Parse: !alert BTC 50000
	parts := strings.Fields(ctx.Message)
	if len(parts) < 3 {
		Reply(ctx, "Usage: !alert <crypto> <price>\nExample: !alert BTC 50000")
		return nil
	}

	crypto := strings.ToUpper(parts[1])
	targetPrice := parts[2]

	price := 0.10

	// Check if agent can afford this
	canSpend, reason := ctx.Agent.CanSpend(price)
	if !canSpend {
		Reply(ctx, fmt.Sprintf("❌ Cannot create alert: %s", reason))
		return nil
	}

	// Create payment invoice first
	Reply(ctx, fmt.Sprintf("💳 Price alert for %s at $%s\n\nThis service costs $%.2f. Please pay with:\n!pay %.2f USDT", 
		crypto, targetPrice, price, price))

	// Note: In production, you'd wait for payment confirmation before setting up alert
	log.Info("Price alert requested", "crypto", crypto, "price", targetPrice, "user", ctx.Sender)

	return nil
}

func (h *AlertHandler) Description() string {
	return "Set price alert for any cryptocurrency"
}

func (h *AlertHandler) Price() float64 {
	return 0.10
}

// SummarizeHandler - Article summarization ($0.50)
type SummarizeHandler struct{}

func (h *SummarizeHandler) Handle(ctx *Context) error {
	// Parse: !summarize <url>
	parts := strings.Fields(ctx.Message)
	if len(parts) < 2 {
		Reply(ctx, "Usage: !summarize <url>\nExample: !summarize https://example.com/article")
		return nil
	}

	url := parts[1]
	price := 0.50

	// Check spending
	canSpend, reason := ctx.Agent.CanSpend(price)
	if !canSpend {
		Reply(ctx, fmt.Sprintf("❌ Cannot summarize: %s", reason))
		return nil
	}

	// Request payment
	Reply(ctx, fmt.Sprintf("📄 Article summarization\nURL: %s\n\nThis service costs $%.2f.\nPay with: !pay %.2f USDT",
		url, price, price))

	// After payment, would:
	// 1. Fetch article
	// 2. Send to OpenAI/LLM
	// 3. Return summary

	log.Info("Summary requested", "url", url, "user", ctx.Sender)
	return nil
}

func (h *SummarizeHandler) Description() string {
	return "Summarize any article or webpage"
}

func (h *SummarizeHandler) Price() float64 {
	return 0.50
}

// ImageHandler - AI image generation ($0.75)
type ImageHandler struct{}

func (h *ImageHandler) Handle(ctx *Context) error {
	// Parse: !image <prompt>
	parts := strings.Fields(ctx.Message)
	if len(parts) < 2 {
		Reply(ctx, "Usage: !image <prompt>\nExample: !image a cat wearing a spacesuit on the moon")
		return nil
	}

	prompt := strings.Join(parts[1:], " ")
	price := 0.75

	canSpend, reason := ctx.Agent.CanSpend(price)
	if !canSpend {
		Reply(ctx, fmt.Sprintf("❌ Cannot generate image: %s", reason))
		return nil
	}

	Reply(ctx, fmt.Sprintf("🎨 AI Image Generation\nPrompt: %s\n\nThis service costs $%.2f.\nPay with: !pay %.2f USDT",
		prompt, price, price))

	log.Info("Image generation requested", "prompt", prompt, "user", ctx.Sender)
	return nil
}

func (h *ImageHandler) Description() string {
	return "Generate AI images from text prompts"
}

func (h *ImageHandler) Price() float64 {
	return 0.75
}

// CodeHandler - Code generation ($0.50)
type CodeHandler struct{}

func (h *CodeHandler) Handle(ctx *Context) error {
	parts := strings.Fields(ctx.Message)
	if len(parts) < 2 {
		Reply(ctx, "Usage: !code <description>\nExample: !code a Python function to calculate fibonacci")
		return nil
	}

	description := strings.Join(parts[1:], " ")
	price := 0.50

	canSpend, reason := ctx.Agent.CanSpend(price)
	if !canSpend {
		Reply(ctx, fmt.Sprintf("❌ Cannot generate code: %s", reason))
		return nil
	}

	Reply(ctx, fmt.Sprintf("💻 Code Generation\nDescription: %s\n\nThis service costs $%.2f.\nPay with: !pay %.2f USDT",
		description, price, price))

	log.Info("Code generation requested", "description", description, "user", ctx.Sender)
	return nil
}

func (h *CodeHandler) Description() string {
	return "Generate code snippets from description"
}

func (h *CodeHandler) Price() float64 {
	return 0.50
}

// ProposeHandler - Agent proposes custom service
type ProposeHandler struct{}

func (h *ProposeHandler) Handle(ctx *Context) error {
	parts := strings.Fields(ctx.Message)
	if len(parts) < 2 {
		Reply(ctx, "Usage: !propose <your idea>\nExample: !propose I need a Python script to scrape prices from Amazon")
		return nil
	}

	idea := strings.Join(parts[1:], " ")

	// Get AI pricing recommendation
	price, reasoning, err := ctx.Agent.DecideServicePricing(nil, idea)
	if err != nil {
		price = 5.00 // Default
		reasoning = "Based on complexity estimate"
	}

	msg := fmt.Sprintf(`🤖 **Custom Service Proposal**

Your request: %s

**Recommended price:** $%.2f
**Reasoning:** %s

Would you like me to proceed? Reply:
• "yes" to confirm and receive payment instructions
• "no" to cancel
• Or suggest a different price`,
		idea, price, reasoning)

	ReplyWithHTML(ctx, msg)

	log.Info("Custom service proposed", "idea", idea, "price", price, "user", ctx.Sender)
	return nil
}

func (h *ProposeHandler) Description() string {
	return "Agent proposes custom service pricing"
}

func (h *ProposeHandler) Price() float64 {
	return -1 // Variable
}

// ServicesHandler lists all services
type ServicesHandler struct{}

func (h *ServicesHandler) Handle(ctx *Context) error {
	services := `📋 **Available Services**

**Free:**
• !help - Show commands
• !balance - Check treasury
• !services - This message
• !price <crypto> - Get crypto prices

**Paid Services:**
• !alert <crypto> <price> - $0.10
  Set price alerts for any cryptocurrency

• !summarize <url> - $0.50
  Get AI summary of any article

• !image <prompt> - $0.75
  Generate AI images (DALL-E style)

• !code <description> - $0.50
  Generate code in any language

• !propose <idea> - $0.50-$1.00
  Custom service proposal

💡 **My limits:** $1 per transaction, $5 per day

All payments in USDT or USDC. Type !pay to send payment.`

	ReplyWithHTML(ctx, services)
	return nil
}

func (h *ServicesHandler) Description() string {
	return "List all available services"
}

func (h *ServicesHandler) Price() float64 {
	return 0
}

// PriceHandler gets crypto prices
type PriceHandler struct{}

func (h *PriceHandler) Handle(ctx *Context) error {
	parts := strings.Fields(ctx.Message)
	if len(parts) < 2 {
		Reply(ctx, "Usage: !price <crypto>\nExample: !price BTC")
		return nil
	}

	crypto := strings.ToUpper(parts[1])
	
	// TODO: Integrate with CoinGecko/CoinMarketCap API
	// For now, placeholder
	Reply(ctx, fmt.Sprintf("💰 **%s Price**\n\nCurrent: $XX,XXX.XX\n24h Change: +X.XX%%\n\n(Powered by CoinGecko)", crypto))
	
	return nil
}

func (h *PriceHandler) Description() string {
	return "Get cryptocurrency prices"
}

func (h *PriceHandler) Price() float64 {
	return 0
}
