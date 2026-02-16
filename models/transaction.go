package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/yanaatere/expense_tracking/internal/db"
)

type Transaction = db.Transaction

type TransactionModel struct {
	q *db.Queries
}

func NewTransactionModel(d db.DBTX) *TransactionModel {
	return &TransactionModel{q: db.New(d)}
}

func (m *TransactionModel) Create(ctx context.Context, userID int32, tType string, amount float64, description string, categoryID *int32, date pgtype.Date) (*Transaction, error) {
	catID := pgtype.Int4{Valid: false}
	if categoryID != nil {
		catID = pgtype.Int4{Int32: *categoryID, Valid: true}
	}

	amountNumeric := pgtype.Numeric{}
	amountNumeric.Scan(amount) // Basic conversion, assuming float64 fits

	t, err := m.q.CreateTransaction(ctx, db.CreateTransactionParams{
		UserID:          userID,
		Type:            tType,
		Amount:          amountNumeric,
		Description:     pgtype.Text{String: description, Valid: true},
		CategoryID:      catID,
		TransactionDate: date,
	})
	if err != nil {
		return nil, err
	}
	return &t, nil
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

func (m *TransactionModel) Update(ctx context.Context, id int32, userID int32, tType string, amount float64, description string, categoryID *int32, date pgtype.Date) (*Transaction, error) {
	catID := pgtype.Int4{Valid: false}
	if categoryID != nil {
		catID = pgtype.Int4{Int32: *categoryID, Valid: true}
	}

	amountNumeric := pgtype.Numeric{}
	amountNumeric.Scan(amount)

	t, err := m.q.UpdateTransaction(ctx, db.UpdateTransactionParams{
		ID:              id,
		UserID:          userID,
		Type:            tType,
		Amount:          amountNumeric,
		Description:     pgtype.Text{String: description, Valid: true},
		CategoryID:      catID,
		TransactionDate: date,
	})
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (m *TransactionModel) Delete(ctx context.Context, id int32, userID int32) error {
	return m.q.DeleteTransaction(ctx, db.DeleteTransactionParams{
		ID:     id,
		UserID: userID,
	})
}

// Stats result
type DashboardStats struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	TotalBalance float64 `json:"total_balance"`
}

func (m *TransactionModel) GetDashboardStats(ctx context.Context, userID int32) (*DashboardStats, error) {
	res, err := m.q.GetDashboardStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert numeric to float64
	income, _ := res.TotalIncome.Float64Value()
	expense, _ := res.TotalExpense.Float64Value()

	return &DashboardStats{
		TotalIncome:  income.Float64,
		TotalExpense: expense.Float64,
		TotalBalance: income.Float64 - expense.Float64,
	}, nil
}
