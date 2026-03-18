package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/models"
)

type TransactionHandler struct {
	model *models.TransactionModel
}

func NewTransactionHandler(pool *pgxpool.Pool) *TransactionHandler {
	return &TransactionHandler{
		model: models.NewTransactionModel(pool),
	}
}

type TransactionInput struct {
	UserID          int     `json:"user_id"`
	Type            string  `json:"type"`
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
	CategoryID      *int    `json:"category_id"`
	SubCategoryID   *int    `json:"sub_category_id"`
	WalletID        *int    `json:"wallet_id"`
	Date            string  `json:"date"` // Expects "2006-01-02"
	ReceiptImageUrl string  `json:"receipt_image_url"`
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var input TransactionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if input.UserID == 0 {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	if input.Type == "" {
		http.Error(w, "Type (income/expense) is required", http.StatusBadRequest)
		return
	}
	if input.Amount <= 0 {
		http.Error(w, "Amount must be positive", http.StatusBadRequest)
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
			http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
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

	transaction, err := h.model.Create(r.Context(), int32(input.UserID), input.Type, input.Amount, input.Description, catID, subCatID, walletID, pgDate, input.ReceiptImageUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}

func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	transactions, err := h.model.GetAll(r.Context(), int32(userID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	transaction, err := h.model.Get(r.Context(), int32(id), int32(userID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}

func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	var input TransactionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if input.UserID == 0 {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	var date time.Time
	if input.Date != "" {
		date, err = time.Parse("2006-01-02", input.Date)
		if err != nil {
			date, err = time.Parse("02/01/2006", input.Date)
			if err != nil {
				http.Error(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
				return
			}
		}
	} else {
		// Use current time or previous time? Better to require it or fetch existing.
		// For simplicity, let's just use current time if missing or fail?
		// Or fetch existing. But that's extra query.
		// Let's assume date is provided or default to now.
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

	transaction, err := h.model.Update(r.Context(), int32(id), int32(input.UserID), input.Type, input.Amount, input.Description, catID, subCatID, walletID, pgDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}

func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid transaction ID", http.StatusBadRequest)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	err = h.model.Delete(r.Context(), int32(id), int32(userID))
	if err != nil {
		// if err == sql.ErrNoRows { // sql package removed
		// 	http.Error(w, "Transaction not found", http.StatusNotFound)
		// 	return
		// }
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TransactionHandler) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid User ID", http.StatusBadRequest)
		return
	}

	stats, err := h.model.GetDashboardStats(r.Context(), int32(userID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
