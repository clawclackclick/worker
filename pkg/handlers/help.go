package handlers

// HelpHandler shows available commands
type HelpHandler struct{}

func (h *HelpHandler) Handle(ctx *Context) error {
 help := `🤖 ClawClack Agent Help

Free Commands:
• !help - Show this message
• !balance - Check my treasury and spending limits
• !services - List all available services
• !price <crypto> - Get current crypto price

Paid Services:
• !alert <crypto> <price> - Set price alert ($0.10)
• !summarize <url> - Summarize any article ($0.50)
• !image <prompt> - Generate AI image ($0.75)
• !code <description> - Generate code snippet ($0.50)
• !propose <idea> - I propose a custom service ($0.50-$1.00)

Payment:
• !pay <amount> <currency> - Send me money
• !status <invoice_id> - Check payment status

My limits: $1/transaction, $5/day

Need something else? Just ask!`

 Reply(ctx, help)
 return nil
}

func (h *HelpHandler) Description() string {
 return "Show help message"
}

func (h *HelpHandler) Price() float64 {
 return 0
}
