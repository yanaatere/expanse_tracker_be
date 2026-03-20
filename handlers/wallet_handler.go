package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/auth"
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
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
	Goals    *string `json:"goals"`
}

// @Summary Get wallets
// @Description Get all wallets for the authenticated user (protected)
// @Tags Wallets
// @Produce json
// @Success 200 {array} object
// @Failure 401 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/wallets [get]
func (h *WalletHandler) GetWallets(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	wallets, err := h.model.GetAll(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, wallets)
}

// @Summary Get wallet
// @Description Get wallet by id for the authenticated user (protected)
// @Tags Wallets
// @Produce json
// @Param id path int true "Wallet ID"
// @Success 200 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 401 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/wallets/{id} [get]
func (h *WalletHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid wallet ID")
		return
	}

	wallet, err := h.model.Get(r.Context(), int32(id), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, wallet)
}

// @Summary Create wallet
// @Description Create wallet for the authenticated user (protected). Required: name, type, currency, balance. Optional: goals.
// @Tags Wallets
// @Accept json
// @Produce json
// @Param request body WalletInput true "Create wallet request"
// @Success 201 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 401 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/wallets [post]
func (h *WalletHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var input WalletInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	if input.Name == "" || input.Type == "" || input.Currency == "" {
		WriteError(w, http.StatusBadRequest, "name, type, and currency are required")
		return
	}

	wallet, err := h.model.Create(r.Context(), userID, input.Name, input.Type, input.Currency, input.Balance, input.Goals)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusCreated, wallet)
}

// @Summary Update wallet
// @Description Update wallet details for the authenticated user (protected). Required: name, type, currency, balance. Optional: goals.
// @Tags Wallets
// @Accept json
// @Produce json
// @Param id path int true "Wallet ID"
// @Param request body WalletInput true "Update wallet request"
// @Success 200 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 401 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/wallets/{id} [put]
func (h *WalletHandler) UpdateWallet(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

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
	if input.Name == "" || input.Type == "" || input.Currency == "" {
		WriteError(w, http.StatusBadRequest, "name, type, and currency are required")
		return
	}

	wallet, err := h.model.Update(r.Context(), int32(id), userID, input.Name, input.Type, input.Currency, input.Balance, input.Goals)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, wallet)
}

// @Summary Get wallet names by type
// @Description Returns predefined wallet names for a given type (bank, e-wallet, cash). Public endpoint.
// @Tags Wallets
// @Produce json
// @Param type query string true "Wallet type: bank, e-wallet, or cash"
// @Success 200 {array} string
// @Router /api/wallets/names [get]
func (h *WalletHandler) GetWalletNames(w http.ResponseWriter, r *http.Request) {
	walletType := strings.ToLower(r.URL.Query().Get("type"))

	var names []string
	bankNames := []string{"Mandiri", "BCA", "BNI", "BRI", "CIMB Niaga", "Danamon", "Permata", "BTN", "Other Bank"}
	switch walletType {
	case "bank", "credit":
		names = bankNames
	case "e-wallet":
		names = []string{"GoPay", "OVO", "Dana", "ShopeePay", "LinkAja", "Other"}
	default:
		names = []string{}
	}

	WriteSuccess(w, http.StatusOK, names)
}

// @Summary Delete wallet
// @Description Delete a wallet by id for the authenticated user (protected)
// @Tags Wallets
// @Param id path int true "Wallet ID"
// @Success 204 {object} MessageResponse
// @Failure 400 {object} MessageResponse
// @Failure 401 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/wallets/{id} [delete]
func (h *WalletHandler) DeleteWallet(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid wallet ID")
		return
	}

	if err := h.model.Delete(r.Context(), int32(id), userID); err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
