package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/yanaatere/expense_tracking/config"
	"github.com/yanaatere/expense_tracking/controllers"
	"github.com/yanaatere/expense_tracking/logger"
	"github.com/yanaatere/expense_tracking/middleware"
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
	balanceController := controllers.NewBalanceController(cfg.DB)
	walletController := controllers.NewWalletController(cfg.DB) // cfg.DB is *pgxpool.Pool

	// Register routes
	userController.RegisterRoutes(r)
	categoryController.RegisterRoutes(r)
	transactionController.RegisterRoutes(r)
	balanceController.RegisterRoutes(r)
	walletController.RegisterRoutes(r)

	// Apply middleware to all routes (order matters: logging -> CORS)
	// Logging middleware must be first to capture all requests
	handler := middleware.LoggingMiddleware(r)
	handler = middleware.CORSMiddleware(handler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Infof("======================== Server Starting ========================")
	logger.Infof("Server starting on port %s", port)
	logger.Infof("Environment: %s", os.Getenv("ENVIRONMENT"))
	logger.Infof("Database: %s", os.Getenv("DB_NAME"))
	logger.Infof("===========================================================")

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}
