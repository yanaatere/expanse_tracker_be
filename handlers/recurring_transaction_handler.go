package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/logger"
	"github.com/yanaatere/expense_tracking/models"
)

type RecurringTransactionHandler struct {
	model RecurringTransactionModelInterface
}

func NewRecurringTransactionHandler(pool *pgxpool.Pool) *RecurringTransactionHandler {
	return &RecurringTransactionHandler{
		model: models.NewRecurringTransactionModel(pool),
	}
}

func NewRecurringTransactionHandlerWithModel(model RecurringTransactionModelInterface) *RecurringTransactionHandler {
	return &RecurringTransactionHandler{model: model}
}

type RecurringTransactionInput struct {
	Title         string  `json:"title"`
	Type          string  `json:"type"`
	Amount        float64 `json:"amount"`
	CategoryID    *int    `json:"category_id"`
	SubCategoryID *int    `json:"sub_category_id"`
	WalletID      *int    `json:"wallet_id"`
	Frequency     string  `json:"frequency"`
	StartDate     string  `json:"start_date"` // "YYYY-MM-DD"
	EndDate       string  `json:"end_date"`   // "YYYY-MM-DD", optional
}

func parseOptionalDate(s string) (pgtype.Date, error) {
	if s == "" {
		return pgtype.Date{Valid: false}, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Date{}, err
	}
	return pgtype.Date{Time: t, Valid: true}, nil
}

// @Summary Create recurring transaction
// @Description Create a new recurring transaction schedule (protected)
// @Tags RecurringTransactions
// @Accept json
// @Produce json
// @Param request body RecurringTransactionInput true "Recurring transaction input"
// @Success 201 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/recurring-transactions [post]
func (h *RecurringTransactionHandler) CreateRecurringTransaction(w http.ResponseWriter, r *http.Request) {
	var input RecurringTransactionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	if input.Title == "" {
		WriteError(w, http.StatusBadRequest, "Title is required")
		return
	}
	if input.Type == "" {
		WriteError(w, http.StatusBadRequest, "Type (income/expense) is required")
		return
	}
	if input.Amount <= 0 {
		WriteError(w, http.StatusBadRequest, "Amount must be positive")
		return
	}
	if input.Frequency == "" {
		WriteError(w, http.StatusBadRequest, "Frequency is required")
		return
	}
	if input.StartDate == "" {
		WriteError(w, http.StatusBadRequest, "Start date is required")
		return
	}

	startDate, err := parseOptionalDate(input.StartDate)
	if err != nil || !startDate.Valid {
		WriteError(w, http.StatusBadRequest, "Invalid start_date. Use YYYY-MM-DD")
		return
	}
	endDate, err := parseOptionalDate(input.EndDate)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid end_date. Use YYYY-MM-DD")
		return
	}

	var catID *int32
	if input.CategoryID != nil {
		v := int32(*input.CategoryID)
		catID = &v
	}
	var subCatID *int32
	if input.SubCategoryID != nil {
		v := int32(*input.SubCategoryID)
		subCatID = &v
	}
	var wID *int32
	if input.WalletID != nil {
		v := int32(*input.WalletID)
		wID = &v
	}

	rt, err := h.model.Create(r.Context(), userID, input.Title, input.Type, input.Amount, catID, subCatID, wID, input.Frequency, startDate, endDate)
	if err != nil {
		logger.Errorf("CreateRecurringTransaction: userID=%d err=%v", userID, err)
		WriteError(w, http.StatusInternalServerError, "Failed to create recurring transaction")
		return
	}

	WriteSuccess(w, http.StatusCreated, rt)
}

// @Summary List recurring transactions
// @Description Get all recurring transaction schedules for the authenticated user (protected)
// @Tags RecurringTransactions
// @Produce json
// @Success 200 {array} object
// @Failure 500 {object} MessageResponse
// @Router /api/recurring-transactions [get]
func (h *RecurringTransactionHandler) GetRecurringTransactions(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	list, err := h.model.GetAll(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to retrieve recurring transactions")
		return
	}

	WriteSuccess(w, http.StatusOK, list)
}

// @Summary Get recurring transaction
// @Description Get a recurring transaction schedule by ID (protected)
// @Tags RecurringTransactions
// @Produce json
// @Param id path int true "Recurring Transaction ID"
// @Success 200 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/recurring-transactions/{id} [get]
func (h *RecurringTransactionHandler) GetRecurringTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid recurring transaction ID")
		return
	}

	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	rt, err := h.model.Get(r.Context(), int32(id), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to retrieve recurring transaction")
		return
	}

	WriteSuccess(w, http.StatusOK, rt)
}

// @Summary Update recurring transaction
// @Description Update a recurring transaction schedule (protected)
// @Tags RecurringTransactions
// @Accept json
// @Produce json
// @Param id path int true "Recurring Transaction ID"
// @Param request body RecurringTransactionInput true "Recurring transaction update input"
// @Success 200 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/recurring-transactions/{id} [put]
func (h *RecurringTransactionHandler) UpdateRecurringTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid recurring transaction ID")
		return
	}

	var input RecurringTransactionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	startDate, err := parseOptionalDate(input.StartDate)
	if err != nil || !startDate.Valid {
		WriteError(w, http.StatusBadRequest, "Invalid start_date. Use YYYY-MM-DD")
		return
	}
	endDate, err := parseOptionalDate(input.EndDate)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid end_date. Use YYYY-MM-DD")
		return
	}

	var catID *int32
	if input.CategoryID != nil {
		v := int32(*input.CategoryID)
		catID = &v
	}
	var subCatID *int32
	if input.SubCategoryID != nil {
		v := int32(*input.SubCategoryID)
		subCatID = &v
	}
	var wID *int32
	if input.WalletID != nil {
		v := int32(*input.WalletID)
		wID = &v
	}

	rt, err := h.model.Update(r.Context(), int32(id), userID, input.Title, input.Type, input.Amount, catID, subCatID, wID, input.Frequency, startDate, endDate, startDate)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to update recurring transaction")
		return
	}

	WriteSuccess(w, http.StatusOK, rt)
}

// @Summary Delete recurring transaction
// @Description Delete a recurring transaction schedule (protected)
// @Tags RecurringTransactions
// @Param id path int true "Recurring Transaction ID"
// @Success 204
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/recurring-transactions/{id} [delete]
func (h *RecurringTransactionHandler) DeleteRecurringTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid recurring transaction ID")
		return
	}

	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.model.Delete(r.Context(), int32(id), userID); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to delete recurring transaction")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
