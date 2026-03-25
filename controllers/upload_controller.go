package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/handlers"
)

type UploadController struct {
	handler *handlers.UploadHandler
}

func NewUploadController() *UploadController {
	return &UploadController{
		handler: handlers.NewUploadHandler(),
	}
}

func (c *UploadController) RegisterRoutes(router *mux.Router) {
	router.Handle(
		"/api/uploads/receipts",
		auth.JWTMiddleware(http.HandlerFunc(c.handler.UploadReceipt)),
	).Methods("POST")
}
