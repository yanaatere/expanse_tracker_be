package models

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/internal/db"
)

type RecurringTransaction = db.RecurringTransaction

type RecurringTransactionModel struct {
	q                *db.Queries
	pool             *pgxpool.Pool
	transactionModel *TransactionModel
}

func NewRecurringTransactionModel(pool *pgxpool.Pool) *RecurringTransactionModel {
	return &RecurringTransactionModel{
		q:                db.New(pool),
		pool:             pool,
		transactionModel: NewTransactionModel(pool),
	}
}

func (m *RecurringTransactionModel) Create(
	ctx context.Context,
	userID int32,
	title, tType string,
	amount float64,
	categoryID, subCategoryID, walletID *int32,
	frequency string,
	startDate, endDate pgtype.Date,
) (*RecurringTransaction, error) {
	amountNumeric := pgtype.Numeric{}
	if err := amountNumeric.Scan(strconv.FormatFloat(amount, 'f', -1, 64)); err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

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

	row, err := m.q.CreateRecurringTransaction(ctx, db.CreateRecurringTransactionParams{
		UserID:            userID,
		Title:             title,
		Type:              tType,
		Amount:            amountNumeric,
		CategoryID:        catID,
		SubCategoryID:     subCatID,
		WalletID:          wID,
		Frequency:         frequency,
		StartDate:         startDate,
		EndDate:           endDate,
		NextExecutionDate: startDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create recurring transaction: %w", err)
	}

	return &row, nil
}

func (m *RecurringTransactionModel) GetAll(ctx context.Context, userID int32) ([]RecurringTransaction, error) {
	rows, err := m.q.ListRecurringTransactions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list recurring transactions: %w", err)
	}
	return rows, nil
}

func (m *RecurringTransactionModel) Get(ctx context.Context, id, userID int32) (*RecurringTransaction, error) {
	row, err := m.q.GetRecurringTransaction(ctx, db.GetRecurringTransactionParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get recurring transaction: %w", err)
	}
	return &row, nil
}

func (m *RecurringTransactionModel) Update(
	ctx context.Context,
	id, userID int32,
	title, tType string,
	amount float64,
	categoryID, subCategoryID, walletID *int32,
	frequency string,
	startDate, endDate, nextExecutionDate pgtype.Date,
) (*RecurringTransaction, error) {
	amountNumeric := pgtype.Numeric{}
	if err := amountNumeric.Scan(strconv.FormatFloat(amount, 'f', -1, 64)); err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

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

	row, err := m.q.UpdateRecurringTransaction(ctx, db.UpdateRecurringTransactionParams{
		ID:                id,
		UserID:            userID,
		Title:             title,
		Type:              tType,
		Amount:            amountNumeric,
		CategoryID:        catID,
		SubCategoryID:     subCatID,
		WalletID:          wID,
		Frequency:         frequency,
		StartDate:         startDate,
		EndDate:           endDate,
		NextExecutionDate: nextExecutionDate,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update recurring transaction: %w", err)
	}

	return &row, nil
}

func (m *RecurringTransactionModel) Delete(ctx context.Context, id, userID int32) error {
	err := m.q.DeleteRecurringTransaction(ctx, db.DeleteRecurringTransactionParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete recurring transaction: %w", err)
	}
	return nil
}

// ProcessDue queries all active recurring transactions that are due today or earlier,
// creates the corresponding transaction for each, and advances the next execution date.
// Intended to be called by the daily background job in main.go.
func (m *RecurringTransactionModel) ProcessDue(ctx context.Context) error {
	today := pgtype.Date{Time: time.Now(), Valid: true}

	due, err := m.q.ListDueRecurringTransactions(ctx, today)
	if err != nil {
		return fmt.Errorf("failed to list due recurring transactions: %w", err)
	}

	for _, r := range due {
		amountFloat, err := numericToFloat(r.Amount)
		if err != nil {
			continue
		}

		var catID *int32
		if r.CategoryID.Valid {
			v := r.CategoryID.Int32
			catID = &v
		}
		var subCatID *int32
		if r.SubCategoryID.Valid {
			v := r.SubCategoryID.Int32
			subCatID = &v
		}
		var wID *int32
		if r.WalletID.Valid {
			v := r.WalletID.Int32
			wID = &v
		}

		_, err = m.transactionModel.Create(
			ctx,
			r.UserID,
			r.Type,
			amountFloat,
			r.Title,
			catID,
			subCatID,
			wID,
			r.NextExecutionDate,
			"",
		)
		if err != nil {
			// Log and continue — don't stop processing others on a single failure.
			continue
		}

		next := advanceDate(r.NextExecutionDate.Time, r.Frequency)
		nextPg := pgtype.Date{Time: next, Valid: true}

		// Deactivate if past end date.
		if r.EndDate.Valid && next.After(r.EndDate.Time) {
			_ = m.q.DeactivateRecurringTransaction(ctx, r.ID)
			continue
		}

		_, _ = m.q.UpdateNextExecutionDate(ctx, db.UpdateNextExecutionDateParams{
			ID:                r.ID,
			NextExecutionDate: nextPg,
		})
	}

	return nil
}

// advanceDate returns the next execution date based on frequency.
func advanceDate(t time.Time, frequency string) time.Time {
	switch frequency {
	case "daily":
		return t.AddDate(0, 0, 1)
	case "weekly":
		return t.AddDate(0, 0, 7)
	case "monthly":
		return t.AddDate(0, 1, 0)
	case "yearly":
		return t.AddDate(1, 0, 0)
	default:
		return t.AddDate(0, 1, 0)
	}
}

// numericToFloat converts pgtype.Numeric to float64.
func numericToFloat(n pgtype.Numeric) (float64, error) {
	var s string
	if err := n.Scan(&s); err != nil {
		return 0, err
	}
	return strconv.ParseFloat(s, 64)
}
