package handlers

import (
	"context"
	"fmt"

	"github.com/charmbracelet/log"
)

// BalanceHandler shows agent's treasury
type BalanceHandler struct{}

func (h *BalanceHandler) Handle(ctx *Context) error {
	// Get balances from SHKeeper
	balances, err := ctx.SHKeeper.GetBalances(context.Background())
	if err != nil {
		log.Error("Failed to get balances", "error", err)
		Reply(ctx, "⚠️ Unable to fetch balances right now. Try again later.")
		return err
	}

	// Get spending stats from agent
	stats := ctx.Agent.GetSpendingStats()

	msg := "💰 **Agent Treasury**\n\n"
	
	if len(balances) == 0 {
		msg += "No funds available yet.\n"
	} else {
		for currency, amount := range balances {
			msg += fmt.Sprintf("• %s: %s\n", currency, amount)
		}
	}

	msg += fmt.Sprintf("\n📊 **Spending Limits**\n")
	msg += fmt.Sprintf("• Per transaction: $%.2f\n", ctx.Agent.GetSpendingLimit())
	msg += fmt.Sprintf("• Daily budget: $%.2f\n", ctx.Agent.GetDailyBudget())
	msg += fmt.Sprintf("• Spent today: $%.2f\n", stats.SpentToday)
	msg += fmt.Sprintf("• Remaining today: $%.2f\n", ctx.Agent.GetDailyBudget()-stats.SpentToday)

	if stats.LastSpendTime.IsZero() {
		msg += "\n✅ No spending yet today"
	} else {
		msg += fmt.Sprintf("\n🕐 Last spend: %s", stats.LastSpendTime.Format("15:04"))
	}

	ReplyWithHTML(ctx, msg)
	return nil
}

func (h *BalanceHandler) Description() string {
	return "Check agent treasury"
}

func (h *BalanceHandler) Price() float64 {
	return 0
}
