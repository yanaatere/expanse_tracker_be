package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/models"
)

type BalanceHandler struct {
	model *models.BalanceModel
}

func NewBalanceHandler(pool *pgxpool.Pool) *BalanceHandler {
	return &BalanceHandler{
		model: models.NewBalanceModel(pool),
	}
}

// GetBalance returns the overall balance (total income, total expense, total balance) for a user.
// total_income and total_expense are computed from transactions.
// total_balance is read from the balances table.
// GET /api/balance?user_id=1
func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	balance, err := h.model.GetUserBalance(r.Context(), int32(userID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

// GetMonthlyBalance returns the monthly breakdown of income, expense, and net.
// GET /api/balance/monthly?user_id=1
func (h *BalanceHandler) GetMonthlyBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	months, err := h.model.GetMonthlyBalance(r.Context(), int32(userID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(months)
}

// GetBalanceByDateRange returns balance filtered by a date range.
// GET /api/balance/range?user_id=1&start_date=2026-01-01&end_date=2026-01-31
func (h *BalanceHandler) GetBalanceByDateRange(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		http.Error(w, "start_date and end_date are required (format: YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Invalid start_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "Invalid end_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	pgStartDate := pgtype.Date{Time: startDate, Valid: true}
	pgEndDate := pgtype.Date{Time: endDate, Valid: true}

	balance, err := h.model.GetBalanceByDateRange(r.Context(), int32(userID), pgStartDate, pgEndDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

// GetBalanceByCategory returns balance breakdown per category.
// GET /api/balance/category?user_id=1
func (h *BalanceHandler) GetBalanceByCategory(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	categories, err := h.model.GetBalanceByCategory(r.Context(), int32(userID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// RecalculateBalance recalculates the balance from scratch (safety net).
// POST /api/balance/recalculate?user_id=1
func (h *BalanceHandler) RecalculateBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	balance, err := h.model.RecalculateBalance(r.Context(), int32(userID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

// parseUserID extracts and validates user_id from query params
func parseUserID(r *http.Request) (int, error) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		return 0, errUserIDRequired
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, errInvalidUserID
	}
	return userID, nil
}

type balanceError string

func (e balanceError) Error() string { return string(e) }

const (
	errUserIDRequired balanceError = "User ID is required"
	errInvalidUserID  balanceError = "Invalid User ID"
)
