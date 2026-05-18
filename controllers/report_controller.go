package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/handlers"
	"github.com/yanaatere/expense_tracking/middleware"
)

type ReportController struct {
	handler *handlers.ReportHandler
	pool    *pgxpool.Pool
}

func NewReportController(pool *pgxpool.Pool) *ReportController {
	return &ReportController{
		handler: handlers.NewReportHandler(pool),
		pool:    pool,
	}
}

func (c *ReportController) RegisterRoutes(router *mux.Router) {
	premiumChain := func(h http.Handler) http.Handler {
		return auth.JWTMiddleware(middleware.PremiumRequired(c.pool)(h))
	}
	router.Handle(
		"/api/reports/transactions",
		premiumChain(http.HandlerFunc(c.handler.GetReportTransactions)),
	).Methods("GET")
}
