package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/auth"
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
	// All transaction routes are protected - require JWT authentication
	router.Handle("/api/transactions", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetTransactions))).Methods("GET")
	router.Handle("/api/transactions", auth.JWTMiddleware(http.HandlerFunc(c.handler.CreateTransaction))).Methods("POST")
	router.Handle("/api/transactions/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetTransaction))).Methods("GET")
	router.Handle("/api/transactions/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.UpdateTransaction))).Methods("PUT")
	router.Handle("/api/transactions/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.DeleteTransaction))).Methods("DELETE")

	// Wallet transactions (filtered by wallet, optional type filter)
	router.Handle("/api/wallets/{id:[0-9]+}/transactions", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetTransactionsByWallet))).Methods("GET")

	// Dashboard Stats - also protected
	router.Handle("/api/dashboard/stats", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetDashboardStats))).Methods("GET")
}
