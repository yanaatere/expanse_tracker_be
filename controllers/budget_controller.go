package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/handlers"
	"github.com/yanaatere/expense_tracking/internal/db"
)

type BudgetController struct {
	handler *handlers.BudgetHandler
}

func NewBudgetController(d db.DBTX) *BudgetController {
	return &BudgetController{handler: handlers.NewBudgetHandler(d)}
}

func (c *BudgetController) RegisterRoutes(router *mux.Router) {
	router.Handle("/api/budgets", auth.JWTMiddleware(http.HandlerFunc(c.handler.ListBudgets))).Methods("GET")
	router.Handle("/api/budgets", auth.JWTMiddleware(http.HandlerFunc(c.handler.CreateBudget))).Methods("POST")
	router.Handle("/api/budgets/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.UpdateBudget))).Methods("PUT")
	router.Handle("/api/budgets/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.DeleteBudget))).Methods("DELETE")
}
