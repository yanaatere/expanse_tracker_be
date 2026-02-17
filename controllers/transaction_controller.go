package controllers

import (
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/handlers"
)

type TransactionController struct {
	handler *handlers.TransactionHandler
}

func NewTransactionController(pool *pgxpool.Pool) *TransactionController {
	return &TransactionController{
		handler: handlers.NewTransactionHandler(pool),
	}
}

func (c *TransactionController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/transactions", c.handler.GetTransactions).Methods("GET")
	router.HandleFunc("/api/transactions", c.handler.CreateTransaction).Methods("POST")
	router.HandleFunc("/api/transactions/{id:[0-9]+}", c.handler.GetTransaction).Methods("GET")
	router.HandleFunc("/api/transactions/{id:[0-9]+}", c.handler.UpdateTransaction).Methods("PUT")
	router.HandleFunc("/api/transactions/{id:[0-9]+}", c.handler.DeleteTransaction).Methods("DELETE")

	// Dashboard Stats
	router.HandleFunc("/api/dashboard/stats", c.handler.GetDashboardStats).Methods("GET")
}
