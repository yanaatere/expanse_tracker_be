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
	"github.com/yanaatere/expense_tracking/models"
)

type TransactionHandler struct {
	model TransactionModelInterface
}

func NewTransactionHandler(pool *pgxpool.Pool) *TransactionHandler {
	return &TransactionHandler{
		model: models.NewTransactionModel(pool),
	}
}

func NewTransactionHandlerWithModel(model TransactionModelInterface) *TransactionHandler {
	return &TransactionHandler{model: model}
}

type TransactionInput struct {
	Type            string  `json:"type"`
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
	CategoryID      *int    `json:"category_id"`
	SubCategoryID   *int    `json:"sub_category_id"`
	WalletID        *int    `json:"wallet_id"`
	Date            string  `json:"date"` // Expects "2006-01-02"
	ReceiptImageUrl string  `json:"receipt_image_url"`
}

// @Summary Create transaction
// @Description Create a new transaction record (protected)
// @Tags Transactions
// @Accept json
// @Produce json
// @Param request body TransactionInput true "Transaction create request"
// @Success 201 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/transactions [post]
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var input TransactionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
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
	if input.Date == "" {
		input.Date = time.Now().Format("2006-01-02")
	}

	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		// Try parsing DD/MM/YYYY just in case relevant to user locale
		date, err = time.Parse("02/01/2006", input.Date)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
			return
		}
	}

	var catID *int32
	if input.CategoryID != nil {
		id := int32(*input.CategoryID)
		catID = &id
	}
	var subCatID *int32
	if input.SubCategoryID != nil {
		id := int32(*input.SubCategoryID)
		subCatID = &id
	}
	var walletID *int32
	if input.WalletID != nil {
		id := int32(*input.WalletID)
		walletID = &id
	}

	pgDate := pgtype.Date{
		Time:  date,
		Valid: true,
	}

	transaction, err := h.model.Create(r.Context(), userID, input.Type, input.Amount, input.Description, catID, subCatID, walletID, pgDate, input.ReceiptImageUrl)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusCreated, transaction)
}

// @Summary Get transactions
// @Description Get all transactions for a user (protected)
// @Tags Transactions
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {array} object
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/transactions [get]
func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	transactions, err := h.model.GetAll(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, transactions)
}

// @Summary Get transaction
// @Description Get transaction by id for a user (protected)
// @Tags Transactions
// @Produce json
// @Param id path int true "Transaction ID"
// @Param user_id query int true "User ID"
// @Success 200 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/transactions/{id} [get]
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid transaction ID")
		return
	}

	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	transaction, err := h.model.Get(r.Context(), int32(id), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, transaction)
}

// @Summary Update transaction
// @Description Update transaction record by id (protected)
// @Tags Transactions
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Param request body TransactionInput true "Transaction update request"
// @Success 200 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/transactions/{id} [put]
func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid transaction ID")
		return
	}

	var input TransactionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var date time.Time
	if input.Date != "" {
		date, err = time.Parse("2006-01-02", input.Date)
		if err != nil {
			date, err = time.Parse("02/01/2006", input.Date)
			if err != nil {
				WriteError(w, http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
				return
			}
		}
	} else {
		date = time.Now()
	}

	var catID *int32
	if input.CategoryID != nil {
		cid := int32(*input.CategoryID)
		catID = &cid
	}
	var subCatID *int32
	if input.SubCategoryID != nil {
		sid := int32(*input.SubCategoryID)
		subCatID = &sid
	}
	var walletID *int32
	if input.WalletID != nil {
		wid := int32(*input.WalletID)
		walletID = &wid
	}

	pgDate := pgtype.Date{
		Time:  date,
		Valid: true,
	}

	transaction, err := h.model.Update(r.Context(), int32(id), userID, input.Type, input.Amount, input.Description, catID, subCatID, walletID, pgDate)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, transaction)
}

// @Summary Delete transaction
// @Description Delete transaction by id for a user (protected)
// @Tags Transactions
// @Param id path int true "Transaction ID"
// @Param user_id query int true "User ID"
// @Success 204 {object} MessageResponse
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/transactions/{id} [delete]
func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid transaction ID")
		return
	}

	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	err = h.model.Delete(r.Context(), int32(id), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Get transactions by wallet
// @Description Get all transactions for a specific wallet, with optional type filter (protected)
// @Tags Transactions
// @Produce json
// @Param id path int true "Wallet ID"
// @Param type query string false "Filter by type: income or expense"
// @Success 200 {array} object
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/wallets/{id}/transactions [get]
func (h *TransactionHandler) GetTransactionsByWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	walletID, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid wallet ID")
		return
	}

	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	typeFilter := r.URL.Query().Get("type")

	transactions, err := h.model.GetByWallet(r.Context(), userID, int32(walletID), typeFilter)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, transactions)
}

// @Summary Get dashboard stats
// @Description Get transaction dashboard stats for a user (protected)
// @Tags Transactions
// @Produce json
// @Param user_id query int true "User ID"
// @Success 200 {object} interface{}
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/dashboard/stats [get]
func (h *TransactionHandler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	stats, err := h.model.GetDashboardStats(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, stats)
}
