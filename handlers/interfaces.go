package handlers

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/yanaatere/expense_tracking/internal/db"
	"github.com/yanaatere/expense_tracking/models"
)

// UserModelInterface abstracts user database operations for testability.
type UserModelInterface interface {
	GetAll(ctx context.Context) ([]models.User, error)
	Get(ctx context.Context, id int32) (*models.User, error)
	Create(ctx context.Context, username, email string) (*models.User, error)
	CreateWithPassword(ctx context.Context, username, email, password string) (*models.User, error)
	Update(ctx context.Context, id int32, username, email string) (*models.User, error)
	Delete(ctx context.Context, id int32) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	UpdatePassword(ctx context.Context, id int32, hashedPassword string) (*models.User, error)
	SetPasswordResetToken(ctx context.Context, id int32, token string, expiresAt time.Time) (*models.User, error)
	GetByResetToken(ctx context.Context, token string) (*models.User, error)
	ClearPasswordResetToken(ctx context.Context, id int32) (*models.User, error)
}

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

// WalletModelInterface abstracts wallet database operations for testability.
type WalletModelInterface interface {
	GetAll(ctx context.Context, userID int32) ([]models.Wallet, error)
	Get(ctx context.Context, id, userID int32) (*models.Wallet, error)
	Create(ctx context.Context, userID int32, name, walletType, currency string, balance float64, goals *string, backdropImage *string) (*models.Wallet, error)
	Update(ctx context.Context, id, userID int32, name, walletType, currency string, balance float64, goals *string, backdropImage *string) (*models.Wallet, error)
	Delete(ctx context.Context, id, userID int32) error
}

// RecurringTransactionModelInterface abstracts recurring transaction operations for testability.
type RecurringTransactionModelInterface interface {
	Create(ctx context.Context, userID int32, title, tType string, amount float64, categoryID, subCategoryID, walletID *int32, frequency string, startDate, endDate pgtype.Date) (*models.RecurringTransaction, error)
	GetAll(ctx context.Context, userID int32) ([]models.RecurringTransaction, error)
	Get(ctx context.Context, id, userID int32) (*models.RecurringTransaction, error)
	Update(ctx context.Context, id, userID int32, title, tType string, amount float64, categoryID, subCategoryID, walletID *int32, frequency string, startDate, endDate, nextExecutionDate pgtype.Date) (*models.RecurringTransaction, error)
	Delete(ctx context.Context, id, userID int32) error
}

// BalanceModelInterface abstracts balance database operations for testability.
type BalanceModelInterface interface {
	GetUserBalance(ctx context.Context, userID int32) (*models.UserBalanceResponse, error)
	GetMonthlyBalance(ctx context.Context, userID int32) ([]models.MonthlyBalance, error)
	GetBalanceByDateRange(ctx context.Context, userID int32, startDate, endDate pgtype.Date) (*models.UserBalanceResponse, error)
	RecalculateBalance(ctx context.Context, userID int32) (*models.UserBalanceResponse, error)
	GetHomeSummary(ctx context.Context, userID int32, loc *time.Location) (*models.HomeSummaryResponse, error)
}
