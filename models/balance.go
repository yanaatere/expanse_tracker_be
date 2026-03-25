package models

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/internal/db"
)

type BalanceModel struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewBalanceModel(pool *pgxpool.Pool) *BalanceModel {
	return &BalanceModel{
		q:    db.New(pool),
		pool: pool,
	}
}

// UserBalanceResponse is the combined response for the balance endpoint
type UserBalanceResponse struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
	TotalBalance float64 `json:"total_balance"`
}

// MonthlyBalance represents income/expense/net for a specific month
type MonthlyBalance struct {
	Month   string  `json:"month"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Net     float64 `json:"net"`
}

// CategoryBalance represents total amount per category per type
type CategoryBalance struct {
	CategoryID       *int32  `json:"category_id"`
	CategoryName     string  `json:"category_name"`
	Type             string  `json:"type"`
	TotalAmount      float64 `json:"total_amount"`
	TransactionCount int64   `json:"transaction_count"`
}

// GetUserBalance returns the combined balance:
// - total_income and total_expense are computed from transactions
// - total_balance is read from the balances table
func (m *BalanceModel) GetUserBalance(ctx context.Context, userID int32) (*UserBalanceResponse, error) {
	// Get income/expense from transactions (computed)
	stats, err := m.q.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	income, _ := stats.TotalIncome.Float64Value()
	expense, _ := stats.TotalExpense.Float64Value()

	// Get balance from balances table
	bal, err := m.q.GetBalance(ctx, userID)
	if err != nil {
		if err == pgx.ErrNoRows {
			// No balance row yet — return 0
			return &UserBalanceResponse{
				TotalIncome:  income.Float64,
				TotalExpense: expense.Float64,
				TotalBalance: 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	balance, _ := bal.TotalBalance.Float64Value()

	return &UserBalanceResponse{
		TotalIncome:  income.Float64,
		TotalExpense: expense.Float64,
		TotalBalance: balance.Float64,
	}, nil
}

// AdjustBalance atomically adjusts the balance by a delta amount.
// Positive delta for income, negative delta for expense.
func (m *BalanceModel) AdjustBalance(ctx context.Context, userID int32, delta float64) error {
	deltaNumeric := pgtype.Numeric{}
	if err := deltaNumeric.Scan(strconv.FormatFloat(delta, 'f', -1, 64)); err != nil {
		return fmt.Errorf("invalid delta: %w", err)
	}

	_, err := m.q.AdjustBalance(ctx, db.AdjustBalanceParams{
		UserID:       userID,
		TotalBalance: deltaNumeric,
	})
	return err
}

// AdjustBalanceWithTx adjusts the balance within an existing transaction
func (m *BalanceModel) AdjustBalanceWithTx(ctx context.Context, tx pgx.Tx, userID int32, delta float64) error {
	qtx := m.q.WithTx(tx)

	deltaNumeric := pgtype.Numeric{}
	if err := deltaNumeric.Scan(strconv.FormatFloat(delta, 'f', -1, 64)); err != nil {
		return fmt.Errorf("invalid delta: %w", err)
	}

	_, err := qtx.AdjustBalance(ctx, db.AdjustBalanceParams{
		UserID:       userID,
		TotalBalance: deltaNumeric,
	})
	return err
}

// RecalculateBalance recalculates the balance from all transactions.
// Useful as a safety net or admin endpoint.
func (m *BalanceModel) RecalculateBalance(ctx context.Context, userID int32) (*UserBalanceResponse, error) {
	bal, err := m.q.RecalculateBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to recalculate balance: %w", err)
	}

	// Also get income/expense
	stats, err := m.q.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	income, _ := stats.TotalIncome.Float64Value()
	expense, _ := stats.TotalExpense.Float64Value()
	balance, _ := bal.TotalBalance.Float64Value()

	return &UserBalanceResponse{
		TotalIncome:  income.Float64,
		TotalExpense: expense.Float64,
		TotalBalance: balance.Float64,
	}, nil
}

func (m *BalanceModel) GetMonthlyBalance(ctx context.Context, userID int32) ([]MonthlyBalance, error) {
	rows, err := m.q.GetMonthlyBalance(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []MonthlyBalance
	for _, row := range rows {
		income, _ := row.Income.Float64Value()
		expense, _ := row.Expense.Float64Value()
		net, _ := row.Net.Float64Value()

		month := ""
		if row.Month.Valid {
			month = row.Month.Time.Format("2006-01")
		}

		result = append(result, MonthlyBalance{
			Month:   month,
			Income:  income.Float64,
			Expense: expense.Float64,
			Net:     net.Float64,
		})
	}
	return result, nil
}

func (m *BalanceModel) GetBalanceByDateRange(ctx context.Context, userID int32, startDate, endDate pgtype.Date) (*UserBalanceResponse, error) {
	res, err := m.q.GetBalanceByDateRange(ctx, db.GetBalanceByDateRangeParams{
		UserID:            userID,
		TransactionDate:   startDate,
		TransactionDate_2: endDate,
	})
	if err != nil {
		return nil, err
	}

	income, _ := res.TotalIncome.Float64Value()
	expense, _ := res.TotalExpense.Float64Value()
	balance, _ := res.Balance.Float64Value()

	return &UserBalanceResponse{
		TotalIncome:  income.Float64,
		TotalExpense: expense.Float64,
		TotalBalance: balance.Float64,
	}, nil
}

func (m *BalanceModel) GetBalanceByCategory(ctx context.Context, userID int32) ([]CategoryBalance, error) {
	rows, err := m.q.GetBalanceByCategory(ctx, userID)
	if err != nil {
		return nil, err
	}

	var result []CategoryBalance
	for _, row := range rows {
		amount, _ := row.TotalAmount.Float64Value()

		var catID *int32
		if row.CategoryID.Valid {
			catID = &row.CategoryID.Int32
		}

		catName := "Uncategorized"
		if row.CategoryName.Valid {
			catName = row.CategoryName.String
		}

		result = append(result, CategoryBalance{
			CategoryID:       catID,
			CategoryName:     catName,
			Type:             row.Type,
			TotalAmount:      amount.Float64,
			TransactionCount: row.TransactionCount,
		})
	}
	return result, nil
}

// BeginTx starts a new database transaction
func (m *BalanceModel) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return m.pool.Begin(ctx)
}
