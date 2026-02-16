package models

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/yanaatere/expense_tracking/internal/db"
)

type Category = db.Category

type CategoryModel struct {
	q *db.Queries
}

func NewCategoryModel(d db.DBTX) *CategoryModel {
	return &CategoryModel{q: db.New(d)}
}

func (m *CategoryModel) GetAll(ctx context.Context) ([]Category, error) {
	return m.q.ListCategories(ctx)
}

func (m *CategoryModel) Get(ctx context.Context, id int32) (*Category, error) {
	c, err := m.q.GetCategory(ctx, id)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (m *CategoryModel) Create(ctx context.Context, name, description string) (*Category, error) {
	c, err := m.q.CreateCategory(ctx, db.CreateCategoryParams{
		Name: name,
		Description: pgtype.Text{
			String: description,
			Valid:  true,
		},
	})
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (m *CategoryModel) Update(ctx context.Context, id int32, name, description string) (*Category, error) {
	c, err := m.q.UpdateCategory(ctx, db.UpdateCategoryParams{
		ID:   id,
		Name: name,
		Description: pgtype.Text{
			String: description,
			Valid:  true,
		},
	})
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (m *CategoryModel) Delete(ctx context.Context, id int32) error {
	return m.q.DeleteCategory(ctx, id)
}
