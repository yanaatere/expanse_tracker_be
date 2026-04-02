package models

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/internal/db"
)

type Transaction = db.Transaction

type TransactionModel struct {
	q            *db.Queries
	pool         *pgxpool.Pool
	balanceModel *BalanceModel
	walletModel  *WalletModel
}

func NewTransactionModel(pool *pgxpool.Pool) *TransactionModel {
	return &TransactionModel{
		q:            db.New(pool),
		pool:         pool,
		balanceModel: NewBalanceModel(pool),
		walletModel:  NewWalletModel(pool),
	}
}

// balanceDelta calculates how much to adjust the balance.
// Income adds to balance (+amount), expense subtracts from balance (-amount).
func balanceDelta(tType string, amount float64) float64 {
	if tType == "income" {
		return amount
	}
	return -amount
}

func (m *TransactionModel) Create(ctx context.Context, userID int32, tType string, amount float64, description string, categoryID *int32, subCategoryID *int32, walletID *int32, date pgtype.Date, receiptImageUrl string) (*Transaction, error) {
	// Start a DB transaction for atomicity
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := m.q.WithTx(tx)

	catID := pgtype.Int4{Valid: false}
	if categoryID != nil {
		catID = pgtype.Int4{Int32: *categoryID, Valid: true}
	}
	subCatID := pgtype.Int4{Valid: false}
	if subCategoryID != nil {
		subCatID = pgtype.Int4{Int32: *subCategoryID, Valid: true}
	}
	wID := pgtype.Int4{Valid: false}
	if walletID != nil {
		wID = pgtype.Int4{Int32: *walletID, Valid: true}
	}

	amountNumeric := pgtype.Numeric{}
	if err := amountNumeric.Scan(strconv.FormatFloat(amount, 'f', -1, 64)); err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	imgUrl := pgtype.Text{Valid: false}
	if receiptImageUrl != "" {
		imgUrl = pgtype.Text{String: receiptImageUrl, Valid: true}
	}

	t, err := qtx.CreateTransaction(ctx, db.CreateTransactionParams{
		UserID:          userID,
		Type:            tType,
		Amount:          amountNumeric,
		Description:     pgtype.Text{String: description, Valid: true},
		CategoryID:      catID,
		SubCategoryID:   subCatID,
		WalletID:        wID,
		TransactionDate: date,
		ReceiptImageUrl: imgUrl,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Adjust global balance
	delta := balanceDelta(tType, amount)
	if err := m.balanceModel.AdjustBalanceWithTx(ctx, tx, userID, delta); err != nil {
		return nil, fmt.Errorf("failed to adjust balance: %w", err)
	}

	// Adjust wallet balance if a wallet is linked
	if walletID != nil {
		if err := m.walletModel.AdjustWalletBalanceWithTx(ctx, tx, *walletID, userID, delta); err != nil {
			return nil, fmt.Errorf("failed to adjust wallet balance: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	result := db.Transaction{ID: t.ID, UserID: t.UserID, Type: t.Type, Amount: t.Amount, Description: t.Description, CategoryID: t.CategoryID, SubCategoryID: t.SubCategoryID, WalletID: t.WalletID, ReceiptImageUrl: t.ReceiptImageUrl, TransactionDate: t.TransactionDate, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt}
	return &result, nil
}

func (m *TransactionModel) GetAll(ctx context.Context, userID int32) ([]db.ListTransactionsRow, error) {
	return m.q.ListTransactions(ctx, userID)
}

func (m *TransactionModel) Get(ctx context.Context, id int32, userID int32) (*db.GetTransactionRow, error) {
	t, err := m.q.GetTransaction(ctx, db.GetTransactionParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (m *TransactionModel) Update(ctx context.Context, id int32, userID int32, tType string, amount float64, description string, categoryID *int32, subCategoryID *int32, walletID *int32, date pgtype.Date) (*Transaction, error) {
	// Start a DB transaction for atomicity
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := m.q.WithTx(tx)

	// First, get the old transaction to reverse its effect on balance
	oldTx, err := qtx.GetTransaction(ctx, db.GetTransactionParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get existing transaction: %w", err)
	}

	// Calculate the old amount as float64
	oldAmountVal, _ := oldTx.Amount.Float64Value()
	oldAmount := oldAmountVal.Float64
	oldDelta := balanceDelta(oldTx.Type, oldAmount)

	catID := pgtype.Int4{Valid: false}
	if categoryID != nil {
		catID = pgtype.Int4{Int32: *categoryID, Valid: true}
	}
	subCatID := pgtype.Int4{Valid: false}
	if subCategoryID != nil {
		subCatID = pgtype.Int4{Int32: *subCategoryID, Valid: true}
	}
	wID := pgtype.Int4{Valid: false}
	if walletID != nil {
		wID = pgtype.Int4{Int32: *walletID, Valid: true}
	}

	amountNumeric := pgtype.Numeric{}
	if err := amountNumeric.Scan(strconv.FormatFloat(amount, 'f', -1, 64)); err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	t, err := qtx.UpdateTransaction(ctx, db.UpdateTransactionParams{
		ID:              id,
		UserID:          userID,
		Type:            tType,
		Amount:          amountNumeric,
		Description:     pgtype.Text{String: description, Valid: true},
		CategoryID:      catID,
		SubCategoryID:   subCatID,
		WalletID:        wID,
		TransactionDate: date,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	// Adjust global balance: reverse old effect, apply new effect
	newDelta := balanceDelta(tType, amount)
	netDelta := newDelta - oldDelta
	if err := m.balanceModel.AdjustBalanceWithTx(ctx, tx, userID, netDelta); err != nil {
		return nil, fmt.Errorf("failed to adjust balance: %w", err)
	}

	// Adjust wallet balances: reverse old wallet, apply new wallet
	if oldTx.WalletID.Valid {
		oldWalletID := oldTx.WalletID.Int32
		if err := m.walletModel.AdjustWalletBalanceWithTx(ctx, tx, oldWalletID, userID, -oldDelta); err != nil {
			return nil, fmt.Errorf("failed to reverse old wallet balance: %w", err)
		}
	}
	if walletID != nil {
		if err := m.walletModel.AdjustWalletBalanceWithTx(ctx, tx, *walletID, userID, newDelta); err != nil {
			return nil, fmt.Errorf("failed to adjust new wallet balance: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}

	result := db.Transaction{ID: t.ID, UserID: t.UserID, Type: t.Type, Amount: t.Amount, Description: t.Description, CategoryID: t.CategoryID, SubCategoryID: t.SubCategoryID, WalletID: t.WalletID, TransactionDate: t.TransactionDate, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt}
	return &result, nil
}

func (m *TransactionModel) Delete(ctx context.Context, id int32, userID int32) error {
	// Start a DB transaction for atomicity
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := m.q.WithTx(tx)

	// First, get the transaction to know how to reverse its balance effect
	oldTx, err := qtx.GetTransaction(ctx, db.GetTransactionParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	// Delete the transaction
	err = qtx.DeleteTransaction(ctx, db.DeleteTransactionParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	// Reverse the global balance effect
	oldAmountVal, _ := oldTx.Amount.Float64Value()
	oldAmount := oldAmountVal.Float64
	reverseDelta := -balanceDelta(oldTx.Type, oldAmount)
	if err := m.balanceModel.AdjustBalanceWithTx(ctx, tx, userID, reverseDelta); err != nil {
		return fmt.Errorf("failed to adjust balance: %w", err)
	}

	// Reverse the wallet balance effect if a wallet was linked
	if oldTx.WalletID.Valid {
		if err := m.walletModel.AdjustWalletBalanceWithTx(ctx, tx, oldTx.WalletID.Int32, userID, reverseDelta); err != nil {
			return fmt.Errorf("failed to reverse wallet balance: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// WalletTransactionRow is the row type returned by GetByWallet.
type WalletTransactionRow struct {
	ID              int32   `json:"id"`
	Type            string  `json:"type"`
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
	CategoryID      *int32  `json:"category_id,omitempty"`
	TransactionDate string  `json:"transaction_date"`
	ReceiptImageUrl string  `json:"receipt_image_url,omitempty"`
}

// GetByWallet returns all transactions for a specific wallet, optionally filtered by type and/or category.
// typeFilter can be "income", "expense", or "" for all.
// categoryID filters by category when non-nil.
func (m *TransactionModel) GetByWallet(ctx context.Context, userID, walletID int32, typeFilter string, categoryID *int32) ([]WalletTransactionRow, error) {
	query := `
		SELECT t.id, t.type, t.amount::float8,
		       COALESCE(t.description, ''),
		       t.category_id,
		       t.transaction_date::text,
		       COALESCE(t.receipt_image_url, '')
		FROM transactions t
		WHERE t.user_id = $1 AND t.wallet_id = $2`

	args := []interface{}{userID, walletID}
	if typeFilter == "income" || typeFilter == "expense" {
		args = append(args, typeFilter)
		query += fmt.Sprintf(` AND t.type = $%d`, len(args))
	}
	if categoryID != nil {
		args = append(args, *categoryID)
		query += fmt.Sprintf(` AND t.category_id = $%d`, len(args))
	}
	query += ` ORDER BY t.transaction_date DESC, t.created_at DESC`

	rows, err := m.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query wallet transactions: %w", err)
	}
	defer rows.Close()

	var result []WalletTransactionRow
	for rows.Next() {
		var r WalletTransactionRow
		if err := rows.Scan(&r.ID, &r.Type, &r.Amount, &r.Description, &r.CategoryID, &r.TransactionDate, &r.ReceiptImageUrl); err != nil {
			return nil, fmt.Errorf("failed to scan wallet transaction: %w", err)
		}
		result = append(result, r)
	}
	if result == nil {
		result = []WalletTransactionRow{}
	}
	return result, rows.Err()
}

// Stats result - kept for backward compatibility with dashboard
type DashboardStats struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	TotalBalance float64 `json:"total_balance"`
}

func (m *TransactionModel) GetDashboardStats(ctx context.Context, userID int32) (*DashboardStats, error) {
	// Get income/expense from transactions
	res, err := m.q.GetDashboardStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	income, _ := res.TotalIncome.Float64Value()
	expense, _ := res.TotalExpense.Float64Value()

	// Get balance from balances table
	bal, err := m.q.GetBalance(ctx, userID)
	if err != nil {
		// If no balance row, compute it
		return &DashboardStats{
			TotalIncome:  income.Float64,
			TotalExpense: expense.Float64,
			TotalBalance: income.Float64 - expense.Float64,
		}, nil
	}

	balance, _ := bal.TotalBalance.Float64Value()

	return &DashboardStats{
		TotalIncome:  income.Float64,
		TotalExpense: expense.Float64,
		TotalBalance: balance.Float64,
	}, nil
}
