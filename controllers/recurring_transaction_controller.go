package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/handlers"
)

type RecurringTransactionController struct {
	handler *handlers.RecurringTransactionHandler
}

func NewRecurringTransactionController(pool *pgxpool.Pool) *RecurringTransactionController {
	return &RecurringTransactionController{
		handler: handlers.NewRecurringTransactionHandler(pool),
	}
}

func (c *RecurringTransactionController) RegisterRoutes(router *mux.Router) {
	router.Handle("/api/recurring-transactions", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetRecurringTransactions))).Methods("GET")
	router.Handle("/api/recurring-transactions", auth.JWTMiddleware(http.HandlerFunc(c.handler.CreateRecurringTransaction))).Methods("POST")
	router.Handle("/api/recurring-transactions/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetRecurringTransaction))).Methods("GET")
	router.Handle("/api/recurring-transactions/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.UpdateRecurringTransaction))).Methods("PUT")
	router.Handle("/api/recurring-transactions/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.DeleteRecurringTransaction))).Methods("DELETE")
}
