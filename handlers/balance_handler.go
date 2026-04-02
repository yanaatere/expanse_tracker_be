package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/models"
)

type BalanceHandler struct {
	model BalanceModelInterface
}

func NewBalanceHandler(pool *pgxpool.Pool) *BalanceHandler {
	return &BalanceHandler{
		model: models.NewBalanceModel(pool),
	}
}

func NewBalanceHandlerWithModel(model BalanceModelInterface) *BalanceHandler {
	return &BalanceHandler{model: model}
}

// GetBalance returns the overall balance (total income, total expense, total balance) for a user.
// total_income and total_expense are computed from transactions.
// total_balance is read from the balances table.
// GET /api/balance?user_id=1
// @Summary Get balance
// @Description Get overall balance for user (protected)
// @Tags Balances
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {object} interface{}
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/balance [get]
func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	balance, err := h.model.GetUserBalance(r.Context(), int32(userID))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, balance)
}

// GetMonthlyBalance returns the monthly breakdown of income, expense, and net.
// GET /api/balance/monthly?user_id=1
// @Summary Get monthly balance
// @Description Get monthly balance breakdown for user (protected)
// @Tags Balances
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {array} interface{}
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/balance/monthly [get]
func (h *BalanceHandler) GetMonthlyBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	months, err := h.model.GetMonthlyBalance(r.Context(), int32(userID))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, months)
}

// GetBalanceByDateRange returns balance filtered by a date range.
// GET /api/balance/range?user_id=1&start_date=2026-01-01&end_date=2026-01-31
// @Summary Get balance by date range
// @Description Get balance summary for date range (protected)
// @Tags Balances
// @Produce json
// @Param user_id query int true "User ID"
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} interface{}
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/balance/range [get]
func (h *BalanceHandler) GetBalanceByDateRange(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		WriteError(w, http.StatusBadRequest, "start_date and end_date are required (format: YYYY-MM-DD)")
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid start_date format. Use YYYY-MM-DD")
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid end_date format. Use YYYY-MM-DD")
		return
	}

	pgStartDate := pgtype.Date{Time: startDate, Valid: true}
	pgEndDate := pgtype.Date{Time: endDate, Valid: true}

	balance, err := h.model.GetBalanceByDateRange(r.Context(), int32(userID), pgStartDate, pgEndDate)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, balance)
}

// RecalculateBalance recalculates the balance from scratch (safety net).
// POST /api/balance/recalculate?user_id=1
// @Summary Recalculate balance
// @Description Recalculate user balance from transactions (protected)
// @Tags Balances
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {object} interface{}
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/balance/recalculate [post]
func (h *BalanceHandler) RecalculateBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := parseUserID(r)
	if err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	balance, err := h.model.RecalculateBalance(r.Context(), int32(userID))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, balance)
}

// GetHomeSummary returns the home screen summary: current month totals,
// previous month expense, percent change vs prev month, and total balance.
// GET /api/home/summary?user_id=1
// @Summary Get home summary
// @Description Get home screen summary (current month totals and percent change) (protected)
// @Tags Balances
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {object} interface{}
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/home/summary [get]
func (h *BalanceHandler)GetHomeSummary(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	if loc == nil {
		loc = time.UTC
	}

	summary, err := h.model.GetHomeSummary(r.Context(), int32(userID), loc)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, summary)
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
