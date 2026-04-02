package models

import (
	"context"
	"fmt"
	"strconv"
	"time"

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

// HomeSummaryResponse is the home screen summary for the current month
type HomeSummaryResponse struct {
	TotalExpense         float64 `json:"total_expense"`
	TotalIncome          float64 `json:"total_income"`
	TotalBalance         float64 `json:"total_balance"`
	PrevMonthExpense     float64 `json:"prev_month_expense"`
	ExpensePercentChange float64 `json:"expense_percent_change"`
	CurrentMonthLabel    string  `json:"current_month_label"`
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

// GetHomeSummary returns the home screen summary: current month totals,
// previous month expense, percent change, and stored total balance.
// loc is the client's timezone (use time.UTC if unknown).
func (m *BalanceModel) GetHomeSummary(ctx context.Context, userID int32, loc *time.Location) (*HomeSummaryResponse, error) {
	now := time.Now().In(loc)

	currStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	currEnd := currStart.AddDate(0, 1, -1)
	prevStart := currStart.AddDate(0, -1, 0)
	prevEnd := currStart.AddDate(0, 0, -1)

	curr, err := m.GetBalanceByDateRange(ctx, userID,
		pgtype.Date{Time: currStart, Valid: true},
		pgtype.Date{Time: currEnd, Valid: true},
	)
	if err != nil {
		return nil, fmt.Errorf("GetHomeSummary: current month: %w", err)
	}

	prev, err := m.GetBalanceByDateRange(ctx, userID,
		pgtype.Date{Time: prevStart, Valid: true},
		pgtype.Date{Time: prevEnd, Valid: true},
	)
	if err != nil {
		return nil, fmt.Errorf("GetHomeSummary: previous month: %w", err)
	}

	// Stored total balance from balances table
	var storedBalance float64
	bal, err := m.q.GetBalance(ctx, userID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, fmt.Errorf("GetHomeSummary: get balance: %w", err)
	}
	if err == nil {
		b, _ := bal.TotalBalance.Float64Value()
		storedBalance = b.Float64
	}

	var pctChange float64
	switch {
	case prev.TotalExpense == 0 && curr.TotalExpense > 0:
		pctChange = 100
	case prev.TotalExpense != 0:
		pctChange = ((curr.TotalExpense - prev.TotalExpense) / prev.TotalExpense) * 100
	}

	return &HomeSummaryResponse{
		TotalExpense:         curr.TotalExpense,
		TotalIncome:          curr.TotalIncome,
		TotalBalance:         storedBalance,
		PrevMonthExpense:     prev.TotalExpense,
		ExpensePercentChange: pctChange,
		CurrentMonthLabel: now.Format("January 2006"),
	}, nil
}

// BeginTx starts a new database transaction
func (m *BalanceModel) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return m.pool.Begin(ctx)
}
