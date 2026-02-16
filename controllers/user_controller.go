package controllers

import (
	"github.com/gorilla/mux"
	"github.com/yanaatere/expense_tracking/handlers"
	"github.com/yanaatere/expense_tracking/internal/db"
)

type UserController struct {
	handler *handlers.UserHandler
}

func NewUserController(d db.DBTX) *UserController {
	return &UserController{
		handler: handlers.NewUserHandler(d),
	}
}

func (c *UserController) RegisterRoutes(router *mux.Router) {
	// User routes
	router.HandleFunc("/api/users", c.handler.GetUsers).Methods("GET")
	router.HandleFunc("/api/users/{id:[0-9]+}", c.handler.GetUser).Methods("GET")
	router.HandleFunc("/api/users", c.handler.CreateUser).Methods("POST")
	router.HandleFunc("/api/users/{id:[0-9]+}", c.handler.UpdateUser).Methods("PUT")
	router.HandleFunc("/api/users/{id:[0-9]+}", c.handler.DeleteUser).Methods("DELETE")
}
