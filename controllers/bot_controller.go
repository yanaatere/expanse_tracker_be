package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/handlers"
)

type BotController struct {
	handler *handlers.BotHandler
}

func NewBotController(redisClient *redis.Client) *BotController {
	return &BotController{
		handler: handlers.NewBotHandler(redisClient),
	}
}

func (c *BotController) RegisterRoutes(router *mux.Router) {
	router.Handle("/api/bot/link", auth.JWTMiddleware(http.HandlerFunc(c.handler.LinkBotUser))).Methods("POST")
}
