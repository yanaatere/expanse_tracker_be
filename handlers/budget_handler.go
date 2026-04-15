package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/internal/db"
	"github.com/yanaatere/expense_tracking/models"
)

type BudgetHandler struct {
	model *models.BudgetModel
}

func NewBudgetHandler(d db.DBTX) *BudgetHandler {
	return &BudgetHandler{model: models.NewBudgetModel(d)}
}

type BudgetInput struct {
	CategoryID          *int32  `json:"category_id"`
	CategoryName        string  `json:"category_name"`
	Limit               float64 `json:"limit"`
	Period              string  `json:"period"`
	Title               *string `json:"title"`
	NotificationEnabled bool    `json:"notification_enabled"`
}

// ListBudgets godoc
// @Summary List budgets
// @Description Get all budgets for the authenticated user
// @Tags Budgets
// @Produce json
// @Security BearerAuth
// @Success 200 {array} object
// @Failure 401 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/budgets [get]
func (h *BudgetHandler) ListBudgets(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	budgets, err := h.model.GetAll(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to retrieve budgets")
		return
	}

	WriteSuccess(w, http.StatusOK, budgets)
}

// CreateBudget godoc
// @Summary Create or upsert a budget
// @Description Create a budget for the authenticated user. Upserts on (user_id, category_name).
// @Tags Budgets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BudgetInput true "Budget input"
// @Success 201 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 401 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/budgets [post]
func (h *BudgetHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input BudgetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if input.CategoryName == "" {
		WriteError(w, http.StatusBadRequest, "category_name is required")
		return
	}
	if input.Period == "" {
		input.Period = "monthly"
	}

	budget, err := h.model.Create(r.Context(), userID, models.BudgetParams{
		CategoryID:          input.CategoryID,
		CategoryName:        input.CategoryName,
		Limit:               input.Limit,
		Period:              input.Period,
		Title:               input.Title,
		NotificationEnabled: input.NotificationEnabled,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create budget")
		return
	}

	WriteSuccess(w, http.StatusCreated, budget)
}

// UpdateBudget godoc
// @Summary Update a budget
// @Description Update an existing budget by ID for the authenticated user
// @Tags Budgets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Budget ID"
// @Param request body BudgetInput true "Budget input"
// @Success 200 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 401 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/budgets/{id} [put]
func (h *BudgetHandler) UpdateBudget(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid budget ID")
		return
	}

	var input BudgetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if input.CategoryName == "" {
		WriteError(w, http.StatusBadRequest, "category_name is required")
		return
	}
	if input.Period == "" {
		input.Period = "monthly"
	}

	budget, err := h.model.Update(r.Context(), int32(id), userID, models.BudgetParams{
		CategoryID:          input.CategoryID,
		CategoryName:        input.CategoryName,
		Limit:               input.Limit,
		Period:              input.Period,
		Title:               input.Title,
		NotificationEnabled: input.NotificationEnabled,
	})
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to update budget")
		return
	}

	WriteSuccess(w, http.StatusOK, budget)
}

// DeleteBudget godoc
// @Summary Delete a budget
// @Description Delete a budget by ID for the authenticated user
// @Tags Budgets
// @Produce json
// @Security BearerAuth
// @Param id path int true "Budget ID"
// @Success 204
// @Failure 400 {object} MessageResponse
// @Failure 401 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/budgets/{id} [delete]
func (h *BudgetHandler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid budget ID")
		return
	}

	if err := h.model.Delete(r.Context(), int32(id), userID); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to delete budget")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
