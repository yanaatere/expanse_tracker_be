package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/handlers"
	"github.com/yanaatere/expense_tracking/internal/db"
)

type CategoryController struct {
	handler *handlers.CategoryHandler
}

func NewCategoryController(db db.DBTX) *CategoryController {
	return &CategoryController{
		handler: handlers.NewCategoryHandler(db),
	}
}

func (c *CategoryController) RegisterRoutes(router *mux.Router) {
	// All category routes are protected - require JWT authentication
	router.Handle("/api/categories", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetCategories))).Methods("GET")
	router.Handle("/api/categories/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetCategory))).Methods("GET")
	router.Handle("/api/categories/{id:[0-9]+}/sub-categories", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetSubCategories))).Methods("GET")
	router.Handle("/api/categories", auth.JWTMiddleware(http.HandlerFunc(c.handler.CreateCategory))).Methods("POST")
	router.Handle("/api/categories/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.UpdateCategory))).Methods("PUT")
	router.Handle("/api/categories/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.DeleteCategory))).Methods("DELETE")
}
