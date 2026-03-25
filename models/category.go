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
	rows, err := m.q.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Category, len(rows))
	for i, r := range rows {
		out[i] = db.Category{ID: r.ID, Name: r.Name, Description: r.Description, ParentID: r.ParentID, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
	}
	return out, nil
}

func (m *CategoryModel) Get(ctx context.Context, id int32) (*Category, error) {
	r, err := m.q.GetCategory(ctx, id)
	if err != nil {
		return nil, err
	}
	c := db.Category{ID: r.ID, Name: r.Name, Description: r.Description, ParentID: r.ParentID, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
	return &c, nil
}

func (m *CategoryModel) GetSubCategories(ctx context.Context, parentID int32) ([]Category, error) {
	rows, err := m.q.ListSubCategories(ctx, pgtype.Int4{Int32: parentID, Valid: true})
	if err != nil {
		return nil, err
	}
	out := make([]Category, len(rows))
	for i, r := range rows {
		out[i] = db.Category{ID: r.ID, Name: r.Name, Description: r.Description, ParentID: r.ParentID, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
	}
	return out, nil
}

func (m *CategoryModel) Create(ctx context.Context, name, description string, parentID *int32) (*Category, error) {
	pid := pgtype.Int4{Valid: false}
	if parentID != nil {
		pid = pgtype.Int4{Int32: *parentID, Valid: true}
	}
	r, err := m.q.CreateCategory(ctx, db.CreateCategoryParams{
		Name:        name,
		Description: pgtype.Text{String: description, Valid: true},
		ParentID:    pid,
	})
	if err != nil {
		return nil, err
	}
	c := db.Category{ID: r.ID, Name: r.Name, Description: r.Description, ParentID: r.ParentID, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
	return &c, nil
}

func (m *CategoryModel) Update(ctx context.Context, id int32, name, description string) (*Category, error) {
	r, err := m.q.UpdateCategory(ctx, db.UpdateCategoryParams{
		ID:          id,
		Name:        name,
		Description: pgtype.Text{String: description, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	c := db.Category{ID: r.ID, Name: r.Name, Description: r.Description, ParentID: r.ParentID, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
	return &c, nil
}

func (m *CategoryModel) Delete(ctx context.Context, id int32) error {
	return m.q.DeleteCategory(ctx, id)
}
