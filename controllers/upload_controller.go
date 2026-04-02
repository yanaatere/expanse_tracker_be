package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go/v7"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/handlers"
)

type UploadController struct {
	handler *handlers.UploadHandler
}

func NewUploadController(minioClient *minio.Client, bucket, minioPublicURL string) *UploadController {
	return &UploadController{
		handler: handlers.NewUploadHandler(minioClient, bucket, minioPublicURL),
	}
}

func (c *UploadController) RegisterRoutes(router *mux.Router) {
	router.Handle(
		"/api/uploads/receipts",
		auth.JWTMiddleware(http.HandlerFunc(c.handler.UploadReceipt)),
	).Methods("POST")
	router.Handle(
		"/api/uploads/receipts/{objectName}",
		auth.JWTMiddleware(http.HandlerFunc(c.handler.DeleteReceipt)),
	).Methods("DELETE")
}
