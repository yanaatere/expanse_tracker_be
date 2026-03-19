package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/models"
)

type WalletHandler struct {
	model WalletModelInterface
}

func NewWalletHandler(pool *pgxpool.Pool) *WalletHandler {
	return &WalletHandler{model: models.NewWalletModel(pool)}
}

func NewWalletHandlerWithModel(model WalletModelInterface) *WalletHandler {
	return &WalletHandler{model: model}
}

type WalletInput struct {
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
}

// @Summary Get wallets
// @Description Get all wallets for a user (protected)
// @Tags Wallets
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {array} object
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/wallets [get]
func (h *WalletHandler) GetWallets(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		WriteError(w, http.StatusBadRequest, "User ID is required")
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid User ID")
		return
	}

	wallets, err := h.model.GetAll(r.Context(), int32(userID))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, wallets)
}

// @Summary Get wallet
// @Description Get wallet by id for a user (protected)
// @Tags Wallets
// @Produce json
// @Param id path int true "Wallet ID"
// @Param user_id query int true "User ID"
// @Success 200 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/wallets/{id} [get]
func (h *WalletHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid wallet ID")
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		WriteError(w, http.StatusBadRequest, "User ID is required")
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid User ID")
		return
	}

	wallet, err := h.model.Get(r.Context(), int32(id), int32(userID))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, wallet)
}

// @Summary Create wallet
// @Description Create wallet for a user (protected)
// @Tags Wallets
// @Accept json
// @Produce json
// @Param request body WalletInput true "Create wallet request"
// @Success 201 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/wallets [post]
func (h *WalletHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	var input WalletInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if input.UserID == 0 || input.Name == "" {
		WriteError(w, http.StatusBadRequest, "user_id and name are required")
		return
	}
	if input.Type == "" {
		input.Type = "general"
	}

	wallet, err := h.model.Create(r.Context(), int32(input.UserID), input.Name, input.Type)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusCreated, wallet)
}

// @Summary Update wallet
// @Description Update wallet details (protected)
// @Tags Wallets
// @Accept json
// @Produce json
// @Param id path int true "Wallet ID"
// @Param request body WalletInput true "Update wallet request"
// @Success 200 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/wallets/{id} [put]
func (h *WalletHandler) UpdateWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid wallet ID")
		return
	}

	var input WalletInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if input.UserID == 0 {
		WriteError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	wallet, err := h.model.Update(r.Context(), int32(id), int32(input.UserID), input.Name, input.Type)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, wallet)
}

// @Summary Delete wallet
// @Description Delete a wallet by id for a user (protected)
// @Tags Wallets
// @Param id path int true "Wallet ID"
// @Param user_id query int true "User ID"
// @Success 204 {object} MessageResponse
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/wallets/{id} [delete]
func (h *WalletHandler) DeleteWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid wallet ID")
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		WriteError(w, http.StatusBadRequest, "User ID is required")
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid User ID")
		return
	}

	if err := h.model.Delete(r.Context(), int32(id), int32(userID)); err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
