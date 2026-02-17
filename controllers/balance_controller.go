package controllers

import (
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
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
	router.HandleFunc("/api/balance", c.handler.GetBalance).Methods("GET")
	router.HandleFunc("/api/balance/monthly", c.handler.GetMonthlyBalance).Methods("GET")
	router.HandleFunc("/api/balance/range", c.handler.GetBalanceByDateRange).Methods("GET")
	router.HandleFunc("/api/balance/category", c.handler.GetBalanceByCategory).Methods("GET")
	router.HandleFunc("/api/balance/recalculate", c.handler.RecalculateBalance).Methods("POST")
}
