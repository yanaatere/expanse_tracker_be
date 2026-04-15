package models

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/yanaatere/expense_tracking/internal/db"
)

type Budget = db.Budget

type BudgetModel struct {
	q *db.Queries
}

func NewBudgetModel(d db.DBTX) *BudgetModel {
	return &BudgetModel{q: db.New(d)}
}

type BudgetParams struct {
	CategoryID          *int32
	CategoryName        string
	Limit               float64
	Period              string
	Title               *string
	NotificationEnabled bool
}

func (m *BudgetModel) GetAll(ctx context.Context, userID int32) ([]Budget, error) {
	return m.q.GetBudgetsByUser(ctx, userID)
}

func (m *BudgetModel) GetByID(ctx context.Context, id, userID int32) (*Budget, error) {
	b, err := m.q.GetBudgetByID(ctx, db.GetBudgetByIDParams{ID: id, UserID: userID})
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (m *BudgetModel) Create(ctx context.Context, userID int32, p BudgetParams) (*Budget, error) {
	limitNumeric, err := toNumeric(p.Limit)
	if err != nil {
		return nil, err
	}
	b, err := m.q.CreateBudget(ctx, db.CreateBudgetParams{
		UserID:              userID,
		CategoryID:          toInt4(p.CategoryID),
		CategoryName:        p.CategoryName,
		BudgetLimit:         limitNumeric,
		Period:              p.Period,
		Title:               toText(p.Title),
		NotificationEnabled: p.NotificationEnabled,
	})
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (m *BudgetModel) Update(ctx context.Context, id, userID int32, p BudgetParams) (*Budget, error) {
	limitNumeric, err := toNumeric(p.Limit)
	if err != nil {
		return nil, err
	}
	b, err := m.q.UpdateBudget(ctx, db.UpdateBudgetParams{
		ID:                  id,
		UserID:              userID,
		CategoryID:          toInt4(p.CategoryID),
		CategoryName:        p.CategoryName,
		BudgetLimit:         limitNumeric,
		Period:              p.Period,
		Title:               toText(p.Title),
		NotificationEnabled: p.NotificationEnabled,
	})
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (m *BudgetModel) Delete(ctx context.Context, id, userID int32) error {
	return m.q.DeleteBudget(ctx, db.DeleteBudgetParams{ID: id, UserID: userID})
}

func toNumeric(f float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	err := n.Scan(strconv.FormatFloat(f, 'f', -1, 64))
	return n, err
}

func toInt4(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

func toText(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *v, Valid: true}
}
