// @title Expense Tracker API
// @version 1.0
// @description Expense Tracker REST API for users, categories, transactions, balances, wallets.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@expense-tracker.local
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
// @schemes http

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/yanaatere/expense_tracking/config"
	"github.com/yanaatere/expense_tracking/controllers"
	_ "github.com/yanaatere/expense_tracking/docs"
	"github.com/yanaatere/expense_tracking/logger"
	"github.com/yanaatere/expense_tracking/middleware"
	"github.com/yanaatere/expense_tracking/models"
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
	userController := controllers.NewUserController(cfg.DB, cfg.Redis)
	transactionController := controllers.NewTransactionController(cfg.DB)
	balanceController := controllers.NewBalanceController(cfg.DB)
	walletController := controllers.NewWalletController(cfg.DB) // cfg.DB is *pgxpool.Pool
	uploadController := controllers.NewUploadController(cfg.Minio, config.MinioBucket)
	botController := controllers.NewBotController(cfg.Redis)
	recurringTransactionController := controllers.NewRecurringTransactionController(cfg.DB)

	// Register routes
	userController.RegisterRoutes(r)
	transactionController.RegisterRoutes(r)
	balanceController.RegisterRoutes(r)
	walletController.RegisterRoutes(r)
	uploadController.RegisterRoutes(r)
	botController.RegisterRoutes(r)
	recurringTransactionController.RegisterRoutes(r)

	// Start daily background job for processing due recurring transactions.
	go func() {
		recurringModel := models.NewRecurringTransactionModel(cfg.DB)
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		// Run once at startup, then every 24 hours.
		if err := recurringModel.ProcessDue(context.Background()); err != nil {
			logger.Infof("recurring transactions processor error: %v", err)
		}
		for range ticker.C {
			if err := recurringModel.ProcessDue(context.Background()); err != nil {
				logger.Infof("recurring transactions processor error: %v", err)
			}
		}
	}()

	// Swagger route (enabled only outside production)
	if os.Getenv("ENVIRONMENT") != "production" {
		r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)
	}

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

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}
