package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/handlers"
)

type BalanceController struct {
	handler *handlers.BalanceHandler
}

func NewBalanceController(pool *pgxpool.Pool) *BalanceController {
	return &BalanceController{
		handler: handlers.NewBalanceHandler(pool),
	}
}

func (c *BalanceController) RegisterRoutes(router *mux.Router) {
	// All balance routes are protected - require JWT authentication
	router.Handle("/api/balance", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetBalance))).Methods("GET")
	router.Handle("/api/balance/monthly", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetMonthlyBalance))).Methods("GET")
	router.Handle("/api/balance/range", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetBalanceByDateRange))).Methods("GET")
router.Handle("/api/balance/recalculate", auth.JWTMiddleware(http.HandlerFunc(c.handler.RecalculateBalance))).Methods("POST")
}
