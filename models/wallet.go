package models

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/internal/db"
)

type Wallet = db.Wallet

type WalletModel struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewWalletModel(pool *pgxpool.Pool) *WalletModel {
	return &WalletModel{q: db.New(pool), pool: pool}
}

// AdjustWalletBalanceWithTx adjusts a wallet's balance within an existing DB transaction.
func (m *WalletModel) AdjustWalletBalanceWithTx(ctx context.Context, tx pgx.Tx, walletID, userID int32, delta float64) error {
	deltaNumeric := pgtype.Numeric{}
	if err := deltaNumeric.Scan(strconv.FormatFloat(delta, 'f', -1, 64)); err != nil {
		return fmt.Errorf("invalid wallet delta: %w", err)
	}
	_, err := m.q.WithTx(tx).AdjustWalletBalance(ctx, db.AdjustWalletBalanceParams{
		Balance: deltaNumeric,
		ID:      walletID,
		UserID:  userID,
	})
	return err
}

func (m *WalletModel) GetAll(ctx context.Context, userID int32) ([]Wallet, error) {
	return m.q.ListWallets(ctx, userID)
}

func (m *WalletModel) Get(ctx context.Context, id, userID int32) (*Wallet, error) {
	w, err := m.q.GetWallet(ctx, db.GetWalletParams{ID: id, UserID: userID})
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (m *WalletModel) Create(ctx context.Context, userID int32, name, walletType, currency string, balance float64, goals *string, backdropImage *string) (*Wallet, error) {
	var goalsText pgtype.Text
	if goals != nil {
		goalsText = pgtype.Text{String: *goals, Valid: true}
	}
	var backdropText pgtype.Text
	if backdropImage != nil {
		backdropText = pgtype.Text{String: *backdropImage, Valid: true}
	}
	var balanceNumeric pgtype.Numeric
	if err := balanceNumeric.Scan(strconv.FormatFloat(balance, 'f', -1, 64)); err != nil {
		return nil, err
	}
	w, err := m.q.CreateWallet(ctx, db.CreateWalletParams{
		UserID:        userID,
		Name:          name,
		Type:          walletType,
		Currency:      currency,
		Balance:       balanceNumeric,
		Goals:         goalsText,
		BackdropImage: backdropText,
	})
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (m *WalletModel) Update(ctx context.Context, id, userID int32, name, walletType, currency string, balance float64, goals *string, backdropImage *string) (*Wallet, error) {
	var goalsText pgtype.Text
	if goals != nil {
		goalsText = pgtype.Text{String: *goals, Valid: true}
	}
	var backdropText pgtype.Text
	if backdropImage != nil {
		backdropText = pgtype.Text{String: *backdropImage, Valid: true}
	}
	var balanceNumeric pgtype.Numeric
	if err := balanceNumeric.Scan(strconv.FormatFloat(balance, 'f', -1, 64)); err != nil {
		return nil, err
	}
	w, err := m.q.UpdateWallet(ctx, db.UpdateWalletParams{
		ID:            id,
		UserID:        userID,
		Name:          name,
		Type:          walletType,
		Currency:      currency,
		Balance:       balanceNumeric,
		Goals:         goalsText,
		BackdropImage: backdropText,
	})
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (m *WalletModel) Delete(ctx context.Context, id, userID int32) error {
	return m.q.DeleteWallet(ctx, db.DeleteWalletParams{ID: id, UserID: userID})
}
