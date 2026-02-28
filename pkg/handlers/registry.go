package handlers

import (
	"context"
	"fmt"
	"strings"

	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/id"

	"clawclack/pkg/agent"
	"clawclack/pkg/shkeeper"
)

// Context holds all dependencies for handlers
type Context struct {
	Client   *mautrix.Client
	RoomID   id.RoomID
	Sender   id.UserID
	Message  string
	SHKeeper *shkeeper.Client
	Agent    *agent.Agent
}

// Handler interface for command handlers
type Handler interface {
	Handle(ctx *Context) error
	Description() string
	Price() float64
}

// Registry holds all command handlers
type Registry struct {
	handlers map[string]Handler
}

func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[string]Handler),
	}
}

func (r *Registry) Register(prefix string, handler Handler) {
	r.handlers[prefix] = handler
}

func (r *Registry) Find(message string) Handler {
	for prefix, handler := range r.handlers {
		if strings.HasPrefix(strings.ToLower(message), prefix) {
			return handler
		}
	}
	return nil
}

func (r *Registry) List() map[string]Handler {
	return r.handlers
}

// Reply helper
func Reply(ctx *Context, message string) {
	content := &mautrix.Content{
		MsgType: event.MsgText,
		Body:    message,
	}
	_, _ = ctx.Client.SendMessage(context.Background(), ctx.RoomID, content)
}

// ReplyWithHTML helper
func ReplyWithHTML(ctx *Context, html string) {
	content := &mautrix.Content{
		MsgType:       event.MsgText,
		Body:          stripHTML(html),
		Format:        event.FormatHTML,
		FormattedBody: html,
	}
	_, _ = ctx.Client.SendMessage(context.Background(), ctx.RoomID, content)
}

func stripHTML(html string) string {
	// Simple HTML stripping for fallback
	return strings.ReplaceAll(strings.ReplaceAll(html, "<b>", ""), "</b>", "")
}
