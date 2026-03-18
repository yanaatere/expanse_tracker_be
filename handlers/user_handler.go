package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
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

// @Summary Get users
// @Description Get all users (protected)
// @Tags Users
// @Produce json
// @Success 200 {array} object
// @Failure 500 {object} MessageResponse
// @Router /api/users [get]
func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.model.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// @Summary Get user
// @Description Get user by id (protected)
// @Tags Users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 404 {object} MessageResponse
// @Router /api/users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.model.Get(r.Context(), int32(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.model.Create(r.Context(), input.Username, input.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// @Summary Update user
// @Description Update user info (protected)
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body UserInput true "Update user request"
// @Success 200 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 404 {object} MessageResponse
// @Router /api/users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var input UserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.model.Update(r.Context(), int32(id), input.Username, input.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// @Summary Delete user
// @Description Delete user by id (protected)
// @Tags Users
// @Param id path int true "User ID"
// @Success 204 {object} MessageResponse
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = h.model.Delete(r.Context(), int32(id))
	// if err == sql.ErrNoRows { // sql package removed
	// 	http.Error(w, "User not found", http.StatusNotFound)
	// 	return
	// }
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
