package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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
		Password: "", // This should not be used, use CreateWithPassword instead
	})
	if err != nil {
		return nil, err
	}
	user := User{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Password:  u.Password,
		IsPremium: u.IsPremium,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
	return &user, nil
}

func (m *UserModel) CreateWithPassword(ctx context.Context, username, email, password string) (*User, error) {
	u, err := m.q.CreateUser(ctx, db.CreateUserParams{
		Username: username,
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, err
	}
	user := User{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Password:  u.Password,
		IsPremium: u.IsPremium,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
	return &user, nil
}

func (m *UserModel) SetPremium(ctx context.Context, userID int32, isPremium bool) (*User, error) {
	u, err := m.q.SetUserPremium(ctx, db.SetUserPremiumParams{
		ID:        userID,
		IsPremium: isPremium,
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

func (m *UserModel) GetByEmail(ctx context.Context, email string) (*User, error) {
	u, err := m.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (m *UserModel) GetByUsername(ctx context.Context, username string) (*User, error) {
	u, err := m.q.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (m *UserModel) UpdatePassword(ctx context.Context, id int32, hashedPassword string) (*User, error) {
	u, err := m.q.UpdatePassword(ctx, db.UpdatePasswordParams{
		ID:       id,
		Password: hashedPassword,
	})
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (m *UserModel) SetPasswordResetToken(ctx context.Context, id int32, token string, expiresAt time.Time) (*User, error) {
	u, err := m.q.SetPasswordResetToken(ctx, db.SetPasswordResetTokenParams{
		ID:                   id,
		PasswordResetToken:   pgtype.Text{String: token, Valid: true},
		PasswordResetExpires: pgtype.Timestamp{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (m *UserModel) GetByResetToken(ctx context.Context, token string) (*User, error) {
	u, err := m.q.GetUserByResetToken(ctx, pgtype.Text{String: token, Valid: true})
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (m *UserModel) ClearPasswordResetToken(ctx context.Context, id int32) (*User, error) {
	u, err := m.q.ClearPasswordResetToken(ctx, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
