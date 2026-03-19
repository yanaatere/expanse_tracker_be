package models

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/yanaatere/expense_tracking/internal/db"
)

type Wallet = db.Wallet

type WalletModel struct {
	q *db.Queries
}

func NewWalletModel(d db.DBTX) *WalletModel {
	return &WalletModel{q: db.New(d)}
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

func (m *WalletModel) Create(ctx context.Context, userID int32, name, walletType, currency string, balance float64, goals *string) (*Wallet, error) {
	var goalsText pgtype.Text
	if goals != nil {
		goalsText = pgtype.Text{String: *goals, Valid: true}
	}
	var balanceNumeric pgtype.Numeric
	if err := balanceNumeric.Scan(strconv.FormatFloat(balance, 'f', -1, 64)); err != nil {
		return nil, err
	}
	w, err := m.q.CreateWallet(ctx, db.CreateWalletParams{
		UserID:   userID,
		Name:     name,
		Type:     walletType,
		Currency: currency,
		Balance:  balanceNumeric,
		Goals:    goalsText,
	})
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (m *WalletModel) Update(ctx context.Context, id, userID int32, name, walletType, currency string, balance float64, goals *string) (*Wallet, error) {
	var goalsText pgtype.Text
	if goals != nil {
		goalsText = pgtype.Text{String: *goals, Valid: true}
	}
	var balanceNumeric pgtype.Numeric
	if err := balanceNumeric.Scan(strconv.FormatFloat(balance, 'f', -1, 64)); err != nil {
		return nil, err
	}
	w, err := m.q.UpdateWallet(ctx, db.UpdateWalletParams{
		ID:       id,
		UserID:   userID,
		Name:     name,
		Type:     walletType,
		Currency: currency,
		Balance:  balanceNumeric,
		Goals:    goalsText,
	})
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (m *WalletModel) Delete(ctx context.Context, id, userID int32) error {
	return m.q.DeleteWallet(ctx, db.DeleteWalletParams{ID: id, UserID: userID})
}
