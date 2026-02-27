package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

// Config for AI agent
type Config struct {
	SpendingLimitUSD float64
	DailyBudgetUSD   float64
	OpenAIKey        string
}

// Agent represents the autonomous AI agent
type Agent struct {
	config        Config
	spendingMutex sync.RWMutex
	dailySpending map[string]float64 // Date string -> amount spent
	lastSpendTime time.Time
	transactions  []Transaction
}

// Transaction records a spend/earn
type Transaction struct {
	ID          string
	Type        string // spend, earn
	Amount      float64
	Currency    string
	Description string
	Timestamp   time.Time
	Approved    bool
}

// SpendingStats for reporting
type SpendingStats struct {
	SpentToday    float64
	SpentTotal    float64
	EarnedToday   float64
	EarnedTotal   float64
	TransactionCount int
	LastSpendTime time.Time
}

// New creates a new AI agent
func New(config Config) *Agent {
	return &Agent{
		config:        config,
		dailySpending: make(map[string]float64),
		transactions:  make([]Transaction, 0),
	}
}

// CanSpend checks if agent can spend amount
func (a *Agent) CanSpend(amount float64) (bool, string) {
	a.spendingMutex.RLock()
	defer a.spendingMutex.RUnlock()

	// Check per-transaction limit
	if amount > a.config.SpendingLimitUSD {
		return false, fmt.Sprintf("Amount $%.2f exceeds per-transaction limit of $%.2f", 
			amount, a.config.SpendingLimitUSD)
	}

	// Check daily budget
	today := time.Now().Format("2006-01-02")
	spentToday := a.dailySpending[today]
	if spentToday+amount > a.config.DailyBudgetUSD {
		return false, fmt.Sprintf("Daily budget exceeded. Spent: $%.2f, Budget: $%.2f, Requested: $%.2f",
			spentToday, a.config.DailyBudgetUSD, amount)
	}

	return true, ""
}

// RecordSpend records a spending transaction
func (a *Agent) RecordSpend(ctx context.Context, amount float64, currency, description string) (*Transaction, error) {
	a.spendingMutex.Lock()
	defer a.spendingMutex.Unlock()

	// Double-check limits
	today := time.Now().Format("2006-01-02")
	spentToday := a.dailySpending[today]
	
	if amount > a.config.SpendingLimitUSD {
		return nil, fmt.Errorf("amount exceeds spending limit")
	}
	
	if spentToday+amount > a.config.DailyBudgetUSD {
		return nil, fmt.Errorf("daily budget exceeded")
	}

	// Record transaction
	tx := Transaction{
		ID:          generateID(),
		Type:        "spend",
		Amount:      amount,
		Currency:    currency,
		Description: description,
		Timestamp:   time.Now(),
		Approved:    true,
	}

	a.transactions = append(a.transactions, tx)
	a.dailySpending[today] += amount
	a.lastSpendTime = tx.Timestamp

	log.Info("💸 Agent spent money", 
		"amount", amount, 
		"currency", currency,
		"description", description,
		"remaining_today", a.config.DailyBudgetUSD-a.dailySpending[today])

	return &tx, nil
}

// RecordEarn records earnings
func (a *Agent) RecordEarn(amount float64, currency, description string) {
	a.spendingMutex.Lock()
	defer a.spendingMutex.Unlock()

	tx := Transaction{
		ID:          generateID(),
		Type:        "earn",
		Amount:      amount,
		Currency:    currency,
		Description: description,
		Timestamp:   time.Now(),
		Approved:    true,
	}

	a.transactions = append(a.transactions, tx)

	log.Info("💰 Agent earned money",
		"amount", amount,
		"currency", currency,
		"description", description)
}

// GetSpendingStats returns current spending statistics
func (a *Agent) GetSpendingStats() SpendingStats {
	a.spendingMutex.RLock()
	defer a.spendingMutex.RUnlock()

	today := time.Now().Format("2006-01-02")
	spentToday := a.dailySpending[today]

	var spentTotal, earnedTotal float64
	for _, tx := range a.transactions {
		if tx.Type == "spend" {
			spentTotal += tx.Amount
		} else if tx.Type == "earn" {
			earnedTotal += tx.Amount
		}
	}

	return SpendingStats{
		SpentToday:       spentToday,
		SpentTotal:       spentTotal,
		EarnedTotal:      earnedTotal,
		TransactionCount: len(a.transactions),
		LastSpendTime:    a.lastSpendTime,
	}
}

// GetSpendingLimit returns per-transaction limit
func (a *Agent) GetSpendingLimit() float64 {
	return a.config.SpendingLimitUSD
}

// GetDailyBudget returns daily budget
func (a *Agent) GetDailyBudget() float64 {
	return a.config.DailyBudgetUSD
}

// DecideServicePricing uses AI to price a custom service
func (a *Agent) DecideServicePricing(ctx context.Context, serviceDescription string) (float64, string, error) {
	// TODO: Integrate with OpenAI to analyze service complexity
	// For now, use simple heuristic
	
	// This is where you'd call OpenAI API to:
	// 1. Analyze the service description
	// 2. Compare to historical pricing
	// 3. Consider current demand
	// 4. Return recommended price and reasoning

	// Placeholder implementation - max $1 due to spending limit
	basePrice := 0.50
	
	// Simple complexity scoring (capped at $1)
	if len(serviceDescription) > 100 {
		basePrice = 1.0
	} else if len(serviceDescription) > 50 {
		basePrice = 0.75
	}

	// Hard cap at spending limit ($1)
	if basePrice > a.config.SpendingLimitUSD {
		basePrice = a.config.SpendingLimitUSD
	}

	reasoning := fmt.Sprintf("Based on service complexity. Max price $%.2f due to agent spending limits.", a.config.SpendingLimitUSD)

	return basePrice, reasoning, nil
}

// generateID creates a simple unique ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
