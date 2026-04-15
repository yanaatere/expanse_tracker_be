package ai

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Service wraps the Anthropic client and exposes domain-specific AI methods.
type Service struct {
	client anthropic.Client
	model  anthropic.Model
}

// NewService creates a new AI Service with the given API key.
// model selects the Claude model to use (e.g. anthropic.ModelClaudeSonnet4_6,
// anthropic.ModelClaudeHaiku4_5, anthropic.ModelClaudeOpus4_6).
// Pass an empty string to use the default (claude-sonnet-4-6).
func NewService(apiKey string, model anthropic.Model) *Service {
	if model == "" {
		model = anthropic.ModelClaudeSonnet4_6
	}
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &Service{client: client, model: model}
}

// ChatMessage represents a single turn in a conversation history.
type ChatMessage struct {
	Role string `json:"role"` // "user" or "ai"
	Text string `json:"text"`
}

// Chat sends a message with conversation history and transaction context, returning the AI reply.
func (s *Service) Chat(ctx context.Context, message string, history []ChatMessage, txContext string) (string, error) {
	messages := buildHistory(history)
	messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(message)))

	resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     s.model,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: "You are Monex AI, a helpful personal finance assistant. " +
				"Answer concisely based on the user's transaction data below. " +
				"Amounts are in Indonesian Rupiah (IDR).\n\n" + txContext},
		},
		Messages: messages,
	})
	if err != nil {
		return "", err
	}
	return resp.Content[0].Text, nil
}

// MonthlyReport generates a friendly narrative summary of the current month's spending.
func (s *Service) MonthlyReport(ctx context.Context, summaryJSON string) (string, error) {
	resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     s.model,
		MaxTokens: 512,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				"Generate a short, friendly 3-5 sentence financial summary for this month's spending. " +
					"Be specific with numbers. Format amounts as Rp X,XXX,XXX.\n\nData: " + summaryJSON,
			)),
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Content[0].Text, nil
}

// BudgetSuggestions returns a JSON array of {category, limit} budget recommendations.
func (s *Service) BudgetSuggestions(ctx context.Context, historyJSON string) (string, error) {
	resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     s.model,
		MaxTokens: 512,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				"Based on these average monthly spending amounts per category, suggest realistic monthly budget limits. " +
					"Return ONLY a JSON array, no explanation: [{\"category\": \"Food and Beverage\", \"limit\": 1500000}]\n\n" +
					"Data: " + historyJSON,
			)),
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Content[0].Text, nil
}

// ScanReceipt extracts transaction details from a base64-encoded JPEG receipt image.
func (s *Service) ScanReceipt(ctx context.Context, base64Image string) (string, error) {
	resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     s.model,
		MaxTokens: 256,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewImageBlockBase64("image/jpeg", base64Image),
				anthropic.NewTextBlock(
					"Extract transaction details from this receipt. "+
						"Return ONLY JSON, no explanation: "+
						"{\"amount\": 50000, \"title\": \"Merchant name\", \"category\": \"Food and Beverage\", \"date\": \"2024-04-09\"}. "+
						"Use null for fields you cannot determine. "+
						"Category must be one of: Bills, Education, Entertainment, Food and Beverage, "+
						"Health, Personal, Transport, Shop, Service, Financial, Vacation, Family and Friends, Pet.",
				),
			),
		},
	})
	if err != nil {
		return "", err
	}
	return resp.Content[0].Text, nil
}

func buildHistory(history []ChatMessage) []anthropic.MessageParam {
	msgs := make([]anthropic.MessageParam, 0, len(history))
	for _, h := range history {
		if h.Role == "ai" {
			msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(h.Text)))
		} else {
			msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(h.Text)))
		}
	}
	return msgs
}
