!status %s to check payment status.`, amount, currency, orderID, invoice.PaymentURL, orderID)

	ReplyWithHTML(ctx, msg)

	// Start monitoring payment in background
	go h.monitorPayment(ctx, orderID)

	return nil
}

func (h *PaymentHandler) monitorPayment(ctx *Context, orderID string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	timeout := time.After(30 * time.Minute)

	for {
		select {
		case <-ticker.C:
			status, err := ctx.SHKeeper.CheckPayment(context.Background(), orderID)
			if err != nil {
				continue
			}

			if status.Status == "confirmed" {
				Reply(ctx, fmt.Sprintf("✅ **Payment confirmed!**\nOrder: %s\nThank you! 🙏", orderID))
				return
			}
		case <-timeout:
			Reply(ctx, fmt.Sprintf("⏰ Payment expired. Order: %s", orderID))
			return
		}
	}
}

func (h *PaymentHandler) Description() string {
	return "Send money to agent"
}

func (h *PaymentHandler) Price() float64 {
	return 0 // User pays, not the service
}

// StatusHandler checks payment status
type StatusHandler struct{}

func (h *StatusHandler) Handle(ctx *Context) error {
	parts := strings.Fields(ctx.Message)
	if len(parts) < 2 {
		Reply(ctx, "Usage: !status <invoice_id>")
		return nil
	}

	orderID := parts[1]

	status, err := ctx.SHKeeper.CheckPayment(context.Background(), orderID)
	if err != nil {
		Reply(ctx, "⚠️ Could not check status. Make sure the ID is correct.")
		return err
	}

	msg := fmt.Sprintf("📋 **Payment Status**\n\nOrder: %s\nStatus: %s", orderID, status.Status)
	
	if status.Status == "confirmed" {
		msg += fmt.Sprintf("\nAmount: %s %s\nReceived at: %s", 
			status.Amount, status.Currency, status.ConfirmedAt.Format("15:04 MST"))
	}

	ReplyWithHTML(ctx, msg)
	return nil
}

func (h *StatusHandler) Description() string {
	return "Check payment status"
}

func (h *StatusHandler) Price() float64 {
	return 0
}
