package handlers

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/yanaatere/expense_tracking/models"
)

// BalanceModelInterface abstracts balance database operations for testability.
type BalanceModelInterface interface {
	GetUserBalance(ctx context.Context, userID int32) (*models.UserBalanceResponse, error)
	GetMonthlyBalance(ctx context.Context, userID int32) ([]models.MonthlyBalance, error)
	GetBalanceByDateRange(ctx context.Context, userID int32, startDate, endDate pgtype.Date) (*models.UserBalanceResponse, error)
	RecalculateBalance(ctx context.Context, userID int32) (*models.UserBalanceResponse, error)
	GetHomeSummary(ctx context.Context, userID int32, loc *time.Location) (*models.HomeSummaryResponse, error)
}
