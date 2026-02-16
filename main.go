package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/yanaatere/expense_tracking/config"
	"github.com/yanaatere/expense_tracking/handlers"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Cannot load config:", err)
	}

	// Initialize router
	r := mux.NewRouter()

	// Initialize handlers
	userHandler := handlers.NewUserHandler(cfg.DB)
	categoryHandler := handlers.NewCategoryHandler(cfg.DB)
	transactionHandler := handlers.NewTransactionHandler(cfg.DB)

	// Routes
	// Users
	r.HandleFunc("/api/users", userHandler.GetUsers).Methods("GET")
	r.HandleFunc("/api/users", userHandler.CreateUser).Methods("POST")
	r.HandleFunc("/api/users/{id}", userHandler.GetUser).Methods("GET")
	r.HandleFunc("/api/users/{id}", userHandler.UpdateUser).Methods("PUT")
	r.HandleFunc("/api/users/{id}", userHandler.DeleteUser).Methods("DELETE")

	// Categories
	r.HandleFunc("/api/categories", categoryHandler.GetCategories).Methods("GET")
	r.HandleFunc("/api/categories", categoryHandler.CreateCategory).Methods("POST")
	r.HandleFunc("/api/categories/{id}", categoryHandler.GetCategory).Methods("GET")
	r.HandleFunc("/api/categories/{id}", categoryHandler.UpdateCategory).Methods("PUT")
	r.HandleFunc("/api/categories/{id}", categoryHandler.DeleteCategory).Methods("DELETE")

	// Transactions
	r.HandleFunc("/api/transactions", transactionHandler.GetTransactions).Methods("GET")
	r.HandleFunc("/api/transactions", transactionHandler.CreateTransaction).Methods("POST")
	r.HandleFunc("/api/transactions/{id}", transactionHandler.GetTransaction).Methods("GET")
	r.HandleFunc("/api/transactions/{id}", transactionHandler.UpdateTransaction).Methods("PUT")
	r.HandleFunc("/api/transactions/{id}", transactionHandler.DeleteTransaction).Methods("DELETE")

	// Dashboard Stats
	r.HandleFunc("/api/dashboard/stats", transactionHandler.GetDashboardStats).Methods("GET")

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
