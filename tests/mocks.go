package tests

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/yanaatere/expense_tracking/handlers"
	"github.com/yanaatere/expense_tracking/internal/db"
	"github.com/yanaatere/expense_tracking/models"
)

// Ensure mocks satisfy the interfaces at compile time.
var _ handlers.UserModelInterface = (*MockUserModel)(nil)
var _ handlers.CategoryModelInterface = (*MockCategoryModel)(nil)
var _ handlers.TransactionModelInterface = (*MockTransactionModel)(nil)
var _ handlers.WalletModelInterface = (*MockWalletModel)(nil)
var _ handlers.BalanceModelInterface = (*MockBalanceModel)(nil)

// MockUserModel is a configurable test double for UserModelInterface.
type MockUserModel struct {
	GetAllFn                  func(ctx context.Context) ([]models.User, error)
	GetFn                     func(ctx context.Context, id int32) (*models.User, error)
	CreateFn                  func(ctx context.Context, username, email string) (*models.User, error)
	CreateWithPasswordFn      func(ctx context.Context, username, email, password string) (*models.User, error)
	UpdateFn                  func(ctx context.Context, id int32, username, email string) (*models.User, error)
	DeleteFn                  func(ctx context.Context, id int32) error
	GetByEmailFn              func(ctx context.Context, email string) (*models.User, error)
	GetByUsernameFn           func(ctx context.Context, username string) (*models.User, error)
	UpdatePasswordFn          func(ctx context.Context, id int32, hashedPassword string) (*models.User, error)
	SetPasswordResetTokenFn   func(ctx context.Context, id int32, token string, expiresAt time.Time) (*models.User, error)
	GetByResetTokenFn         func(ctx context.Context, token string) (*models.User, error)
	ClearPasswordResetTokenFn func(ctx context.Context, id int32) (*models.User, error)
}

func (m *MockUserModel) GetAll(ctx context.Context) ([]models.User, error) {
	return m.GetAllFn(ctx)
}
func (m *MockUserModel) Get(ctx context.Context, id int32) (*models.User, error) {
	return m.GetFn(ctx, id)
}
func (m *MockUserModel) Create(ctx context.Context, username, email string) (*models.User, error) {
	return m.CreateFn(ctx, username, email)
}
func (m *MockUserModel) CreateWithPassword(ctx context.Context, username, email, password string) (*models.User, error) {
	return m.CreateWithPasswordFn(ctx, username, email, password)
}
func (m *MockUserModel) Update(ctx context.Context, id int32, username, email string) (*models.User, error) {
	return m.UpdateFn(ctx, id, username, email)
}
func (m *MockUserModel) Delete(ctx context.Context, id int32) error {
	return m.DeleteFn(ctx, id)
}
func (m *MockUserModel) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.GetByEmailFn(ctx, email)
}
func (m *MockUserModel) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	return m.GetByUsernameFn(ctx, username)
}
func (m *MockUserModel) UpdatePassword(ctx context.Context, id int32, hashedPassword string) (*models.User, error) {
	return m.UpdatePasswordFn(ctx, id, hashedPassword)
}
func (m *MockUserModel) SetPasswordResetToken(ctx context.Context, id int32, token string, expiresAt time.Time) (*models.User, error) {
	return m.SetPasswordResetTokenFn(ctx, id, token, expiresAt)
}
func (m *MockUserModel) GetByResetToken(ctx context.Context, token string) (*models.User, error) {
	return m.GetByResetTokenFn(ctx, token)
}
func (m *MockUserModel) ClearPasswordResetToken(ctx context.Context, id int32) (*models.User, error) {
	return m.ClearPasswordResetTokenFn(ctx, id)
}

// MockCategoryModel is a configurable test double for CategoryModelInterface.
type MockCategoryModel struct {
	GetAllFn           func(ctx context.Context) ([]models.Category, error)
	GetFn              func(ctx context.Context, id int32) (*models.Category, error)
	GetSubCategoriesFn func(ctx context.Context, parentID int32) ([]models.Category, error)
	CreateFn           func(ctx context.Context, name, description string, parentID *int32) (*models.Category, error)
	UpdateFn           func(ctx context.Context, id int32, name, description string) (*models.Category, error)
	DeleteFn           func(ctx context.Context, id int32) error
}

func (m *MockCategoryModel) GetAll(ctx context.Context) ([]models.Category, error) {
	return m.GetAllFn(ctx)
}
func (m *MockCategoryModel) Get(ctx context.Context, id int32) (*models.Category, error) {
	return m.GetFn(ctx, id)
}
func (m *MockCategoryModel) GetSubCategories(ctx context.Context, parentID int32) ([]models.Category, error) {
	return m.GetSubCategoriesFn(ctx, parentID)
}
func (m *MockCategoryModel) Create(ctx context.Context, name, description string, parentID *int32) (*models.Category, error) {
	return m.CreateFn(ctx, name, description, parentID)
}
func (m *MockCategoryModel) Update(ctx context.Context, id int32, name, description string) (*models.Category, error) {
	return m.UpdateFn(ctx, id, name, description)
}
func (m *MockCategoryModel) Delete(ctx context.Context, id int32) error {
	return m.DeleteFn(ctx, id)
}

// MockTransactionModel is a configurable test double for TransactionModelInterface.
type MockTransactionModel struct {
	CreateFn            func(ctx context.Context, userID int32, tType string, amount float64, description string, categoryID *int32, subCategoryID *int32, walletID *int32, date pgtype.Date, receiptImageUrl string) (*models.Transaction, error)
	GetAllFn            func(ctx context.Context, userID int32) ([]db.ListTransactionsRow, error)
	GetFn               func(ctx context.Context, id int32, userID int32) (*db.GetTransactionRow, error)
	UpdateFn            func(ctx context.Context, id int32, userID int32, tType string, amount float64, description string, categoryID *int32, subCategoryID *int32, walletID *int32, date pgtype.Date) (*models.Transaction, error)
	DeleteFn            func(ctx context.Context, id int32, userID int32) error
	GetDashboardStatsFn func(ctx context.Context, userID int32) (*models.DashboardStats, error)
}

func (m *MockTransactionModel) Create(ctx context.Context, userID int32, tType string, amount float64, description string, categoryID *int32, subCategoryID *int32, walletID *int32, date pgtype.Date, receiptImageUrl string) (*models.Transaction, error) {
	return m.CreateFn(ctx, userID, tType, amount, description, categoryID, subCategoryID, walletID, date, receiptImageUrl)
}
func (m *MockTransactionModel) GetAll(ctx context.Context, userID int32) ([]db.ListTransactionsRow, error) {
	return m.GetAllFn(ctx, userID)
}
func (m *MockTransactionModel) Get(ctx context.Context, id int32, userID int32) (*db.GetTransactionRow, error) {
	return m.GetFn(ctx, id, userID)
}
func (m *MockTransactionModel) Update(ctx context.Context, id int32, userID int32, tType string, amount float64, description string, categoryID *int32, subCategoryID *int32, walletID *int32, date pgtype.Date) (*models.Transaction, error) {
	return m.UpdateFn(ctx, id, userID, tType, amount, description, categoryID, subCategoryID, walletID, date)
}
func (m *MockTransactionModel) Delete(ctx context.Context, id int32, userID int32) error {
	return m.DeleteFn(ctx, id, userID)
}
func (m *MockTransactionModel) GetDashboardStats(ctx context.Context, userID int32) (*models.DashboardStats, error) {
	return m.GetDashboardStatsFn(ctx, userID)
}

// MockWalletModel is a configurable test double for WalletModelInterface.
type MockWalletModel struct {
	GetAllFn func(ctx context.Context, userID int32) ([]models.Wallet, error)
	GetFn    func(ctx context.Context, id, userID int32) (*models.Wallet, error)
	CreateFn func(ctx context.Context, userID int32, name, walletType string) (*models.Wallet, error)
	UpdateFn func(ctx context.Context, id, userID int32, name, walletType string) (*models.Wallet, error)
	DeleteFn func(ctx context.Context, id, userID int32) error
}

func (m *MockWalletModel) GetAll(ctx context.Context, userID int32) ([]models.Wallet, error) {
	return m.GetAllFn(ctx, userID)
}
func (m *MockWalletModel) Get(ctx context.Context, id, userID int32) (*models.Wallet, error) {
	return m.GetFn(ctx, id, userID)
}
func (m *MockWalletModel) Create(ctx context.Context, userID int32, name, walletType string) (*models.Wallet, error) {
	return m.CreateFn(ctx, userID, name, walletType)
}
func (m *MockWalletModel) Update(ctx context.Context, id, userID int32, name, walletType string) (*models.Wallet, error) {
	return m.UpdateFn(ctx, id, userID, name, walletType)
}
func (m *MockWalletModel) Delete(ctx context.Context, id, userID int32) error {
	return m.DeleteFn(ctx, id, userID)
}

// MockBalanceModel is a configurable test double for BalanceModelInterface.
type MockBalanceModel struct {
	GetUserBalanceFn        func(ctx context.Context, userID int32) (*models.UserBalanceResponse, error)
	GetMonthlyBalanceFn     func(ctx context.Context, userID int32) ([]models.MonthlyBalance, error)
	GetBalanceByDateRangeFn func(ctx context.Context, userID int32, startDate, endDate pgtype.Date) (*models.UserBalanceResponse, error)
	GetBalanceByCategoryFn  func(ctx context.Context, userID int32) ([]models.CategoryBalance, error)
	RecalculateBalanceFn    func(ctx context.Context, userID int32) (*models.UserBalanceResponse, error)
}

func (m *MockBalanceModel) GetUserBalance(ctx context.Context, userID int32) (*models.UserBalanceResponse, error) {
	return m.GetUserBalanceFn(ctx, userID)
}
func (m *MockBalanceModel) GetMonthlyBalance(ctx context.Context, userID int32) ([]models.MonthlyBalance, error) {
	return m.GetMonthlyBalanceFn(ctx, userID)
}
func (m *MockBalanceModel) GetBalanceByDateRange(ctx context.Context, userID int32, startDate, endDate pgtype.Date) (*models.UserBalanceResponse, error) {
	return m.GetBalanceByDateRangeFn(ctx, userID, startDate, endDate)
}
func (m *MockBalanceModel) GetBalanceByCategory(ctx context.Context, userID int32) ([]models.CategoryBalance, error) {
	return m.GetBalanceByCategoryFn(ctx, userID)
}
func (m *MockBalanceModel) RecalculateBalance(ctx context.Context, userID int32) (*models.UserBalanceResponse, error) {
	return m.RecalculateBalanceFn(ctx, userID)
}
