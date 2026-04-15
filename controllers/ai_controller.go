package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/handlers"
	"github.com/yanaatere/expense_tracking/internal/ai"
)

// AIController registers AI assistant routes.
type AIController struct {
	handler *handlers.AIHandler
}

// NewAIController creates an AIController with the given provider and database pool.
// Construct the provider (Anthropic, Gemini, Qwen…) in main.go and pass it here.
func NewAIController(provider ai.Provider, pool *pgxpool.Pool) *AIController {
	return &AIController{
		handler: handlers.NewAIHandler(provider, pool),
	}
}

// RegisterRoutes registers all AI routes on the provided router.
func (c *AIController) RegisterRoutes(router *mux.Router) {
	router.Handle("/api/ai/chat",
		auth.JWTMiddleware(http.HandlerFunc(c.handler.Chat))).Methods("POST")
	router.Handle("/api/ai/monthly-report",
		auth.JWTMiddleware(http.HandlerFunc(c.handler.MonthlyReport))).Methods("POST")
	router.Handle("/api/ai/budget-suggestions",
		auth.JWTMiddleware(http.HandlerFunc(c.handler.BudgetSuggestions))).Methods("POST")
	router.Handle("/api/ai/scan-receipt",
		auth.JWTMiddleware(http.HandlerFunc(c.handler.ScanReceipt))).Methods("POST")
}
