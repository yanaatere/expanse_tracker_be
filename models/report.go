package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/internal/db"
)

type ReportModel struct {
	q *db.Queries
}

func NewReportModel(pool *pgxpool.Pool) *ReportModel {
	return &ReportModel{q: db.New(pool)}
}

func (m *ReportModel) GetByDateRange(ctx context.Context, userID int32, start, end pgtype.Date) ([]db.GetTransactionsByDateRangeRow, error) {
	return m.q.GetTransactionsByDateRange(ctx, db.GetTransactionsByDateRangeParams{
		UserID:            userID,
		TransactionDate:   start,
		TransactionDate_2: end,
	})
}
