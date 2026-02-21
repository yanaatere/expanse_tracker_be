package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/handlers"
	"github.com/yanaatere/expense_tracking/internal/db"
)

type UserController struct {
	handler     *handlers.UserHandler
	authHandler *handlers.AuthHandler
}

func NewUserController(d db.DBTX) *UserController {
	return &UserController{
		handler:     handlers.NewUserHandler(d),
		authHandler: handlers.NewAuthHandler(d),
	}
}

func (c *UserController) RegisterRoutes(router *mux.Router) {
	// Auth routes (public)
	router.HandleFunc("/api/auth/register", c.authHandler.Register).Methods("POST")
	router.HandleFunc("/api/auth/login", c.authHandler.Login).Methods("POST")
	router.HandleFunc("/api/auth/forgot-password", c.authHandler.ForgotPassword).Methods("POST")
	router.HandleFunc("/api/auth/reset-password", c.authHandler.ResetPassword).Methods("POST")

	// User routes (protected)
	router.Handle("/api/users", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetUsers))).Methods("GET")
	router.Handle("/api/users/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetUser))).Methods("GET")
	router.Handle("/api/users/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.UpdateUser))).Methods("PUT")
	router.Handle("/api/users/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.DeleteUser))).Methods("DELETE")
}
