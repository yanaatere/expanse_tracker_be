package models

import (
	"context"

	"github.com/yanaatere/expense_tracking/internal/db"
)

type User = db.User

type UserModel struct {
	q *db.Queries
}

func NewUserModel(d db.DBTX) *UserModel {
	return &UserModel{q: db.New(d)}
}

func (m *UserModel) GetAll(ctx context.Context) ([]User, error) {
	return m.q.ListUsers(ctx)
}

func (m *UserModel) Get(ctx context.Context, id int32) (*User, error) {
	u, err := m.q.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (m *UserModel) Create(ctx context.Context, username, email string) (*User, error) {
	u, err := m.q.CreateUser(ctx, db.CreateUserParams{
		Username: username,
		Email:    email,
	})
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (m *UserModel) Update(ctx context.Context, id int32, username, email string) (*User, error) {
	u, err := m.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:       id,
		Username: username,
		Email:    email,
	})
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (m *UserModel) Delete(ctx context.Context, id int32) error {
	return m.q.DeleteUser(ctx, id)
}
