package handlers

import (
	"context"
	"time"

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
	SetPremium(ctx context.Context, userID int32, isPremium bool) (*models.User, error)
}
