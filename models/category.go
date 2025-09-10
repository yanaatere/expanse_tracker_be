package models

import (
	"database/sql"
	"time"
)

type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CategoryModel struct {
	db *sql.DB
}

func NewCategoryModel(db *sql.DB) *CategoryModel {
	return &CategoryModel{db: db}
}

func (m *CategoryModel) GetAll() ([]Category, error) {
	rows, err := m.db.Query("SELECT id, name, description, created_at, updated_at FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var c Category
		err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}
	return categories, nil
}

func (m *CategoryModel) Get(id int) (*Category, error) {
	var c Category
	err := m.db.QueryRow("SELECT id, name, description, created_at, updated_at FROM categories WHERE id = ?", id).
		Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (m *CategoryModel) Create(name, description string) (*Category, error) {
	now := time.Now()
	result, err := m.db.Exec("INSERT INTO categories (name, description, created_at, updated_at) VALUES (?, ?, ?, ?)",
		name, description, now, now)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &Category{
		ID:          int(id),
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (m *CategoryModel) Update(id int, name, description string) (*Category, error) {
	now := time.Now()
	_, err := m.db.Exec("UPDATE categories SET name = ?, description = ?, updated_at = ? WHERE id = ?",
		name, description, now, id)
	if err != nil {
		return nil, err
	}

	return &Category{
		ID:          id,
		Name:        name,
		Description: description,
		UpdatedAt:   now,
	}, nil
}

func (m *CategoryModel) Delete(id int) error {
	_, err := m.db.Exec("DELETE FROM categories WHERE id = ?", id)
	return err
}