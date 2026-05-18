package handlers

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/yanaatere/expense_tracking/internal/db"
	"github.com/yanaatere/expense_tracking/models"
)

// TransactionModelInterface abstracts transaction database operations for testability.
type TransactionModelInterface interface {
	Create(ctx context.Context, userID int32, tType string, amount float64, description string, categoryID *int32, subCategoryID *int32, walletID *int32, date pgtype.Date, receiptImageUrl string) (*models.Transaction, error)
	GetAll(ctx context.Context, userID int32) ([]db.ListTransactionsRow, error)
	Get(ctx context.Context, id int32, userID int32) (*db.GetTransactionRow, error)
	Update(ctx context.Context, id int32, userID int32, tType string, amount float64, description string, categoryID *int32, subCategoryID *int32, walletID *int32, date pgtype.Date) (*models.Transaction, error)
	Delete(ctx context.Context, id int32, userID int32) error
	GetDashboardStats(ctx context.Context, userID int32) (*models.DashboardStats, error)
	GetByWallet(ctx context.Context, userID, walletID int32, typeFilter string, categoryID *int32) ([]models.WalletTransactionRow, error)
}
