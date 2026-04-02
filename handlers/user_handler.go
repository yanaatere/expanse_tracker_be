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

type UserHandler struct {
	model UserModelInterface
}

func NewUserHandler(db db.DBTX) *UserHandler {
	return &UserHandler{
		model: models.NewUserModel(db),
	}
}

func NewUserHandlerWithModel(model UserModelInterface) *UserHandler {
	return &UserHandler{model: model}
}

type UserInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

// @Summary Get current user
// @Description Get the authenticated user's profile (protected)
// @Tags Users
// @Produce json
// @Success 200 {object} object
// @Failure 401 {object} MessageResponse
// @Failure 404 {object} MessageResponse
// @Router /api/users [get]
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.model.Get(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}
	if user == nil {
		WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	WriteSuccess(w, http.StatusOK, user)
}

// @Summary Get user
// @Description Get the authenticated user by id — must match JWT identity (protected)
// @Tags Users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 403 {object} MessageResponse
// @Failure 404 {object} MessageResponse
// @Router /api/users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if auth.GetUserIDFromContext(r.Context()) != int32(id) {
		WriteError(w, http.StatusForbidden, "Forbidden")
		return
	}

	user, err := h.model.Get(r.Context(), int32(id))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to retrieve user")
		return
	}
	if user == nil {
		WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	WriteSuccess(w, http.StatusOK, user)
}

// @Summary Create user
// @Description Create a user (protected)
// @Tags Users
// @Accept json
// @Produce json
// @Param request body UserInput true "Create user request"
// @Success 201 {object} object
// @Failure 400 {object} MessageResponse
// @Router /api/users [post]
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input UserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.model.Create(r.Context(), input.Username, input.Email)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	WriteSuccess(w, http.StatusCreated, user)
}

// @Summary Update user
// @Description Update the authenticated user's info — must match JWT identity (protected)
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body UserInput true "Update user request"
// @Success 200 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 403 {object} MessageResponse
// @Failure 404 {object} MessageResponse
// @Router /api/users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if auth.GetUserIDFromContext(r.Context()) != int32(id) {
		WriteError(w, http.StatusForbidden, "Forbidden")
		return
	}

	var input UserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.model.Update(r.Context(), int32(id), input.Username, input.Email)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to update user")
		return
	}
	if user == nil {
		WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	WriteSuccess(w, http.StatusOK, user)
}

// @Summary Delete user
// @Description Delete the authenticated user — must match JWT identity (protected)
// @Tags Users
// @Param id path int true "User ID"
// @Success 204 {object} MessageResponse
// @Failure 400 {object} MessageResponse
// @Failure 403 {object} MessageResponse
// @Router /api/users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if auth.GetUserIDFromContext(r.Context()) != int32(id) {
		WriteError(w, http.StatusForbidden, "Forbidden")
		return
	}

	if err := h.model.Delete(r.Context(), int32(id)); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
