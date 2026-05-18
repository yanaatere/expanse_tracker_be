package handlers

import (
	"context"

	"github.com/yanaatere/expense_tracking/models"
)

// WalletModelInterface abstracts wallet database operations for testability.
type WalletModelInterface interface {
	GetAll(ctx context.Context, userID int32) ([]models.Wallet, error)
	Get(ctx context.Context, id, userID int32) (*models.Wallet, error)
	Create(ctx context.Context, userID int32, name, walletType, currency string, balance float64, goals *string, backdropImage *string) (*models.Wallet, error)
	Update(ctx context.Context, id, userID int32, name, walletType, currency string, balance float64, goals *string, backdropImage *string) (*models.Wallet, error)
	Delete(ctx context.Context, id, userID int32) error
}
