package tests

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/yanaatere/expense_tracking/handlers"
	"github.com/yanaatere/expense_tracking/models"
)

func newBalanceRouter(h *handlers.BalanceHandler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/api/balance", h.GetBalance).Methods(http.MethodGet)
	r.HandleFunc("/api/balance/monthly", h.GetMonthlyBalance).Methods(http.MethodGet)
	r.HandleFunc("/api/balance/range", h.GetBalanceByDateRange).Methods(http.MethodGet)
	r.HandleFunc("/api/balance/category", h.GetBalanceByCategory).Methods(http.MethodGet)
	r.HandleFunc("/api/balance/recalculate", h.RecalculateBalance).Methods(http.MethodPost)
	return r
}

func stubBalanceResponse() *models.UserBalanceResponse {
	return &models.UserBalanceResponse{TotalIncome: 1000, TotalExpense: 400, TotalBalance: 600}
}

// ── GetBalance ────────────────────────────────────────────────────────────────

func TestGetBalance_Success(t *testing.T) {
	mock := &MockBalanceModel{
		GetUserBalanceFn: func(_ context.Context, _ int32) (*models.UserBalanceResponse, error) {
			return stubBalanceResponse(), nil
		},
	}
	h := handlers.NewBalanceHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/balance?user_id=1", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp models.UserBalanceResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.TotalBalance != 600 {
		t.Errorf("expected balance 600, got %f", resp.TotalBalance)
	}
}

func TestGetBalance_MissingUserID(t *testing.T) {
	h := handlers.NewBalanceHandlerWithModel(&MockBalanceModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/balance", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetBalance_InvalidUserID(t *testing.T) {
	h := handlers.NewBalanceHandlerWithModel(&MockBalanceModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/balance?user_id=abc", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetBalance_DBError(t *testing.T) {
	mock := &MockBalanceModel{
		GetUserBalanceFn: func(_ context.Context, _ int32) (*models.UserBalanceResponse, error) {
			return nil, errors.New("db error")
		},
	}
	h := handlers.NewBalanceHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/balance?user_id=1", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── GetMonthlyBalance ─────────────────────────────────────────────────────────

func TestGetMonthlyBalance_Success(t *testing.T) {
	mock := &MockBalanceModel{
		GetMonthlyBalanceFn: func(_ context.Context, _ int32) ([]models.MonthlyBalance, error) {
			return []models.MonthlyBalance{
				{Month: "2026-01", Income: 500, Expense: 200, Net: 300},
				{Month: "2026-02", Income: 600, Expense: 150, Net: 450},
			}, nil
		},
	}
	h := handlers.NewBalanceHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/balance/monthly?user_id=1", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var months []models.MonthlyBalance
	json.NewDecoder(w.Body).Decode(&months)
	if len(months) != 2 {
		t.Errorf("expected 2 months, got %d", len(months))
	}
}

func TestGetMonthlyBalance_MissingUserID(t *testing.T) {
	h := handlers.NewBalanceHandlerWithModel(&MockBalanceModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/balance/monthly", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetMonthlyBalance_DBError(t *testing.T) {
	mock := &MockBalanceModel{
		GetMonthlyBalanceFn: func(_ context.Context, _ int32) ([]models.MonthlyBalance, error) {
			return nil, errors.New("db error")
		},
	}
	h := handlers.NewBalanceHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/balance/monthly?user_id=1", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── GetBalanceByDateRange ─────────────────────────────────────────────────────

func TestGetBalanceByDateRange_Success(t *testing.T) {
	mock := &MockBalanceModel{
		GetBalanceByDateRangeFn: func(_ context.Context, _ int32, _ pgtype.Date, _ pgtype.Date) (*models.UserBalanceResponse, error) {
			return stubBalanceResponse(), nil
		},
	}
	h := handlers.NewBalanceHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/balance/range?user_id=1&start_date=2026-01-01&end_date=2026-01-31", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetBalanceByDateRange_MissingDates(t *testing.T) {
	h := handlers.NewBalanceHandlerWithModel(&MockBalanceModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/balance/range?user_id=1", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetBalanceByDateRange_InvalidStartDate(t *testing.T) {
	h := handlers.NewBalanceHandlerWithModel(&MockBalanceModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/balance/range?user_id=1&start_date=bad&end_date=2026-01-31", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetBalanceByDateRange_InvalidEndDate(t *testing.T) {
	h := handlers.NewBalanceHandlerWithModel(&MockBalanceModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/balance/range?user_id=1&start_date=2026-01-01&end_date=bad", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetBalanceByDateRange_MissingUserID(t *testing.T) {
	h := handlers.NewBalanceHandlerWithModel(&MockBalanceModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/balance/range?start_date=2026-01-01&end_date=2026-01-31", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ── GetBalanceByCategory ──────────────────────────────────────────────────────

func TestGetBalanceByCategory_Success(t *testing.T) {
	mock := &MockBalanceModel{
		GetBalanceByCategoryFn: func(_ context.Context, _ int32) ([]models.CategoryBalance, error) {
			return []models.CategoryBalance{
				{CategoryName: "Food", Type: "expense", TotalAmount: 200, TransactionCount: 5},
			}, nil
		},
	}
	h := handlers.NewBalanceHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/balance/category?user_id=1", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var cats []models.CategoryBalance
	json.NewDecoder(w.Body).Decode(&cats)
	if len(cats) != 1 {
		t.Errorf("expected 1 category, got %d", len(cats))
	}
}

func TestGetBalanceByCategory_MissingUserID(t *testing.T) {
	h := handlers.NewBalanceHandlerWithModel(&MockBalanceModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/balance/category", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetBalanceByCategory_DBError(t *testing.T) {
	mock := &MockBalanceModel{
		GetBalanceByCategoryFn: func(_ context.Context, _ int32) ([]models.CategoryBalance, error) {
			return nil, errors.New("db error")
		},
	}
	h := handlers.NewBalanceHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/balance/category?user_id=1", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── RecalculateBalance ────────────────────────────────────────────────────────

func TestRecalculateBalance_Success(t *testing.T) {
	mock := &MockBalanceModel{
		RecalculateBalanceFn: func(_ context.Context, _ int32) (*models.UserBalanceResponse, error) {
			return stubBalanceResponse(), nil
		},
	}
	h := handlers.NewBalanceHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodPost, "/api/balance/recalculate?user_id=1", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRecalculateBalance_MissingUserID(t *testing.T) {
	h := handlers.NewBalanceHandlerWithModel(&MockBalanceModel{})
	req := httptest.NewRequest(http.MethodPost, "/api/balance/recalculate", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRecalculateBalance_DBError(t *testing.T) {
	mock := &MockBalanceModel{
		RecalculateBalanceFn: func(_ context.Context, _ int32) (*models.UserBalanceResponse, error) {
			return nil, errors.New("db error")
		},
	}
	h := handlers.NewBalanceHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodPost, "/api/balance/recalculate?user_id=1", nil)
	w := httptest.NewRecorder()
	newBalanceRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
