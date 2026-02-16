package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/yanaatere/expense_tracking/config"
	"github.com/yanaatere/expense_tracking/controllers"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Cannot load config:", err)
	}

	// Initialize router
	r := mux.NewRouter()

	// Initialize controllers
	userController := controllers.NewUserController(cfg.DB)
	categoryController := controllers.NewCategoryController(cfg.DB)
	transactionController := controllers.NewTransactionController(cfg.DB)

	// Register routes
	userController.RegisterRoutes(r)
	categoryController.RegisterRoutes(r)
	transactionController.RegisterRoutes(r)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
