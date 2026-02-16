package controllers

import (
	"github.com/gorilla/mux"
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
	router.HandleFunc("/api/categories", c.handler.GetCategories).Methods("GET")
	router.HandleFunc("/api/categories/{id:[0-9]+}", c.handler.GetCategory).Methods("GET")
	router.HandleFunc("/api/categories", c.handler.CreateCategory).Methods("POST")
	router.HandleFunc("/api/categories/{id:[0-9]+}", c.handler.UpdateCategory).Methods("PUT")
	router.HandleFunc("/api/categories/{id:[0-9]+}", c.handler.DeleteCategory).Methods("DELETE")
}
