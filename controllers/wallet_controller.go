package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/handlers"
)

type WalletController struct {
	handler *handlers.WalletHandler
}

func NewWalletController(pool *pgxpool.Pool) *WalletController {
	return &WalletController{
		handler: handlers.NewWalletHandler(pool),
	}
}

func (c *WalletController) RegisterRoutes(router *mux.Router) {
	router.Handle("/api/wallets", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetWallets))).Methods("GET")
	router.Handle("/api/wallets/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.GetWallet))).Methods("GET")
	router.Handle("/api/wallets", auth.JWTMiddleware(http.HandlerFunc(c.handler.CreateWallet))).Methods("POST")
	router.Handle("/api/wallets/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.UpdateWallet))).Methods("PUT")
	router.Handle("/api/wallets/{id:[0-9]+}", auth.JWTMiddleware(http.HandlerFunc(c.handler.DeleteWallet))).Methods("DELETE")
}
