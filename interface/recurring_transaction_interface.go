package handlers

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/yanaatere/expense_tracking/models"
)

// RecurringTransactionModelInterface abstracts recurring transaction operations for testability.
type RecurringTransactionModelInterface interface {
	Create(ctx context.Context, userID int32, title, tType string, amount float64, categoryID, subCategoryID, walletID *int32, frequency string, startDate, endDate pgtype.Date) (*models.RecurringTransaction, error)
	GetAll(ctx context.Context, userID int32) ([]models.RecurringTransaction, error)
	Get(ctx context.Context, id, userID int32) (*models.RecurringTransaction, error)
	Update(ctx context.Context, id, userID int32, title, tType string, amount float64, categoryID, subCategoryID, walletID *int32, frequency string, startDate, endDate, nextExecutionDate pgtype.Date) (*models.RecurringTransaction, error)
	Delete(ctx context.Context, id, userID int32) error
}
