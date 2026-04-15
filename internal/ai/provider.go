package ai

import "context"

// Provider is the interface every LLM backend must implement.
// Add a new provider by creating a struct that satisfies this interface.
type Provider interface {
	Chat(ctx context.Context, message string, history []ChatMessage, txContext string) (string, error)
	MonthlyReport(ctx context.Context, summaryJSON string) (string, error)
	BudgetSuggestions(ctx context.Context, historyJSON string) (string, error)
	ScanReceipt(ctx context.Context, base64Image string) (string, error)
}
