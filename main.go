package main

import (
	"log"
	"net/http"
	"os"

	"github.com/yanaatere/expense_tracking/config"
	"github.com/yanaatere/expense_tracking/handlers"
	"github.com/gorilla/mux"
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

	// Routes
	r.HandleFunc("/api/users", userHandler.GetUsers).Methods("GET")
	r.HandleFunc("/api/users", userHandler.CreateUser).Methods("POST")
	r.HandleFunc("/api/users/{id}", userHandler.GetUser).Methods("GET")
	r.HandleFunc("/api/users/{id}", userHandler.UpdateUser).Methods("PUT")
	r.HandleFunc("/api/users/{id}", userHandler.DeleteUser).Methods("DELETE")

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