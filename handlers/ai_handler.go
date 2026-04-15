package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/internal/ai"
	"github.com/yanaatere/expense_tracking/internal/db"
)

// AIHandler handles AI-powered financial assistant endpoints.
type AIHandler struct {
	provider ai.Provider
	queries  *db.Queries
}

// NewAIHandler creates an AIHandler with the given provider and database pool.
// Pass any ai.Provider implementation (Anthropic, Gemini, Qwen, etc.).
func NewAIHandler(provider ai.Provider, pool *pgxpool.Pool) *AIHandler {
	return &AIHandler{
		provider: provider,
		queries:  db.New(pool),
	}
}

// Chat handles POST /api/ai/chat — conversational financial assistant.
func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input struct {
		Message string           `json:"message"`
		History []ai.ChatMessage `json:"history"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Message == "" {
		WriteError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	since := time.Now().AddDate(0, -3, 0)
	txs, err := h.queries.GetTransactionsForAI(r.Context(), db.GetTransactionsForAIParams{
		UserID:          userID,
		TransactionDate: pgtype.Date{Time: since, Valid: true},
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to load transaction context")
		return
	}

	contextJSON, _ := json.Marshal(txs)
	reply, err := h.provider.Chat(r.Context(), input.Message, input.History, string(contextJSON))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "AI request failed")
		return
	}

	WriteSuccess(w, http.StatusOK, map[string]string{"reply": reply})
}

// MonthlyReport handles POST /api/ai/monthly-report — AI narrative of current month spending.
func (h *AIHandler) MonthlyReport(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	totals, err := h.queries.GetMonthlyCategoryTotals(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to load spending data")
		return
	}

	summaryJSON, _ := json.Marshal(totals)
	report, err := h.provider.MonthlyReport(r.Context(), string(summaryJSON))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "AI request failed")
		return
	}

	WriteSuccess(w, http.StatusOK, map[string]string{"report": report})
}

// BudgetSuggestions handles POST /api/ai/budget-suggestions — recommended budget limits per category.
func (h *AIHandler) BudgetSuggestions(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	avgs, err := h.queries.GetAvgMonthlyCategorySpend(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to load spending history")
		return
	}

	historyJSON, _ := json.Marshal(avgs)
	raw, err := h.provider.BudgetSuggestions(r.Context(), string(historyJSON))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "AI request failed")
		return
	}

	var suggestions []map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &suggestions); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to parse AI response")
		return
	}

	WriteSuccess(w, http.StatusOK, map[string]interface{}{"suggestions": suggestions})
}

// ScanReceipt handles POST /api/ai/scan-receipt — extract transaction data from a receipt image.
func (h *AIHandler) ScanReceipt(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input struct {
		Image string `json:"image"` // base64-encoded JPEG
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Image == "" {
		WriteError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	raw, err := h.provider.ScanReceipt(r.Context(), input.Image)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "AI request failed")
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to parse AI response")
		return
	}

	WriteSuccess(w, http.StatusOK, result)
}
