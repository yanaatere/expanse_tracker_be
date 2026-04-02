package controllers

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/handlers"
	"github.com/yanaatere/expense_tracking/internal/db"
	"github.com/yanaatere/expense_tracking/middleware"
)

type UserController struct {
	handler     *handlers.UserHandler
	authHandler *handlers.AuthHandler
	redis       *redis.Client
}

func NewUserController(d db.DBTX, redisClient *redis.Client) *UserController {
	return &UserController{
		handler:     handlers.NewUserHandler(d),
		authHandler: handlers.NewAuthHandler(d),
		redis:       redisClient,
	}
}

func (c *UserController) RegisterRoutes(router *mux.Router) {
	rl := middleware.RateLimit

	// Auth routes (public) — rate-limited per IP.
	// Login: 10 attempts / minute (brute-force protection).
	router.Handle("/api/auth/login",
		rl(c.redis, "auth:login", 10, time.Minute)(http.HandlerFunc(c.authHandler.Login)),
	).Methods("POST")

	// Register: 5 accounts / hour per IP.
	router.Handle("/api/auth/register",
		rl(c.redis, "auth:register", 5, time.Hour)(http.HandlerFunc(c.authHandler.Register)),
	).Methods("POST")

	// Forgot-password: 3 requests / hour per IP (limits email enumeration).
	router.Handle("/api/auth/forgot-password",
		rl(c.redis, "auth:forgot", 3, time.Hour)(http.HandlerFunc(c.authHandler.ForgotPassword)),
	).Methods("POST")

	// Reset-password: 5 attempts / 15 minutes per IP.
	router.Handle("/api/auth/reset-password",
		rl(c.redis, "auth:reset", 5, 15*time.Minute)(http.HandlerFunc(c.authHandler.ResetPassword)),
	).Methods("POST")

	// User routes (protected)
	router.Handle("/api/users", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetUsers))).Methods("GET")
	router.Handle("/api/users/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetUser))).Methods("GET")
	router.Handle("/api/users/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.UpdateUser))).Methods("PUT")
	router.Handle("/api/users/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.DeleteUser))).Methods("DELETE")
}
