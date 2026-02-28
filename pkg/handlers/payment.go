package handlers

import (
 "context"
 "fmt"
 "strings"
 "time"

 "github.com/charmbracelet/log"
 "github.com/google/uuid"
)

// PaymentHandler handles payment creation
type PaymentHandler struct{}

func (h *PaymentHandler) Handle(ctx *Context) error {
 // Parse: !pay <amount> <currency>
 parts := strings.Fields(ctx.Message)
 if len(parts) < 3 {
  Reply(ctx, "Usage: !pay <amount> <currency>\nExample: !pay 10 USDT")
  return nil
 }

 amount := parts[1]
 currency := strings.ToUpper(parts[2])

 // Validate currency
 validCurrencies := map[string]bool{
  "USDT": true, "USDC": true, "BTC": true, "ETH": true,
 }
 if !validCurrencies[currency] {
  Reply(ctx, fmt.Sprintf("❌ Currency %s not supported. Use: USDT, USDC, BTC, ETH", currency))
  return nil
 }

 // Generate unique order ID
 orderID := uuid.New().String()

 // Create invoice via SHKeeper
 invoice, err := ctx.SHKeeper.CreateInvoice(context.Background(), shkeeper.InvoiceRequest{
  OrderID:  orderID,
  Amount:   amount,
  Currency: currency,
 })
 if err != nil {
  log.Error("Failed to create invoice", "error", err)
  Reply(ctx, "⚠️ Failed to create payment invoice. Please try again.")
  return err
 }

 msg := fmt.Sprintf("💳 Payment Request\n\nAmount: %s %s\nOrder ID: %s\n\nPay here: %s\n\nExpires in 30 minutes", 
  amount, currency, orderID, invoice.PaymentURL)

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
    Reply(ctx, fmt.Sprintf("✅ Payment confirmed!\nOrder: %s\nThank you!", orderID))
    
    // Record earnings for agent
    ctx.Agent.RecordEarn(status.Amount, status.Currency, "Service payment")
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
 return 0
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

 msg := fmt.Sprintf("📋 Payment Status\n\nOrder: %s\nStatus: %s", orderID, status.Status)
 
 if status.Status == "confirmed" {
  msg += fmt.Sprintf("\nAmount: %s %s", status.Amount, status.Currency)
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
