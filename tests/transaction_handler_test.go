package tests

import (
	"context"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/yanaatere/expense_tracking/handlers"
	"github.com/yanaatere/expense_tracking/internal/db"
	"github.com/yanaatere/expense_tracking/models"
)

func newTransactionRouter(h *handlers.TransactionHandler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/api/transactions", h.GetTransactions).Methods(http.MethodGet)
	r.HandleFunc("/api/transactions", h.CreateTransaction).Methods(http.MethodPost)
	r.HandleFunc("/api/transactions/{id}", h.GetTransaction).Methods(http.MethodGet)
	r.HandleFunc("/api/transactions/{id}", h.UpdateTransaction).Methods(http.MethodPut)
	r.HandleFunc("/api/transactions/{id}", h.DeleteTransaction).Methods(http.MethodDelete)
	r.HandleFunc("/api/dashboard/stats", h.GetDashboardStats).Methods(http.MethodGet)
	return r
}

func stubTransaction(id int32) *models.Transaction {
	return &models.Transaction{ID: id, UserID: 1, Type: "expense"}
}

// ── CreateTransaction ─────────────────────────────────────────────────────────

func TestCreateTransaction_Success(t *testing.T) {
	mock := &MockTransactionModel{
		CreateFn: func(_ context.Context, userID int32, tType string, _ float64, _ string, _ *int32, _ *int32, _ *int32, _ pgtype.Date, _ string) (*models.Transaction, error) {
			return stubTransaction(1), nil
		},
	}
	h := handlers.NewTransactionHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]any{
		"user_id": 1, "type": "expense", "amount": 50.0, "date": "2026-03-01",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestCreateTransaction_MissingUserID(t *testing.T) {
	h := handlers.NewTransactionHandlerWithModel(&MockTransactionModel{})
	body, _ := json.Marshal(map[string]any{"type": "expense", "amount": 50.0})
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateTransaction_MissingType(t *testing.T) {
	h := handlers.NewTransactionHandlerWithModel(&MockTransactionModel{})
	body, _ := json.Marshal(map[string]any{"user_id": 1, "amount": 50.0})
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateTransaction_InvalidAmount(t *testing.T) {
	h := handlers.NewTransactionHandlerWithModel(&MockTransactionModel{})
	body, _ := json.Marshal(map[string]any{"user_id": 1, "type": "expense", "amount": -10.0})
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateTransaction_InvalidDate(t *testing.T) {
	h := handlers.NewTransactionHandlerWithModel(&MockTransactionModel{})
	body, _ := json.Marshal(map[string]any{"user_id": 1, "type": "expense", "amount": 50.0, "date": "not-a-date"})
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateTransaction_InvalidJSON(t *testing.T) {
	h := handlers.NewTransactionHandlerWithModel(&MockTransactionModel{})
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", bytes.NewReader([]byte("bad")))
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateTransaction_DBError(t *testing.T) {
	mock := &MockTransactionModel{
		CreateFn: func(_ context.Context, _ int32, _ string, _ float64, _ string, _ *int32, _ *int32, _ *int32, _ pgtype.Date, _ string) (*models.Transaction, error) {
			return nil, errors.New("db error")
		},
	}
	h := handlers.NewTransactionHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]any{"user_id": 1, "type": "expense", "amount": 50.0})
	req := httptest.NewRequest(http.MethodPost, "/api/transactions", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── GetTransactions ───────────────────────────────────────────────────────────

func TestGetTransactions_Success(t *testing.T) {
	mock := &MockTransactionModel{
		GetAllFn: func(_ context.Context, _ int32) ([]db.ListTransactionsRow, error) {
			return []db.ListTransactionsRow{{ID: 1, UserID: 1}}, nil
		},
	}
	h := handlers.NewTransactionHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/transactions?user_id=1", nil)
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetTransactions_MissingUserID(t *testing.T) {
	h := handlers.NewTransactionHandlerWithModel(&MockTransactionModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/transactions", nil)
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetTransactions_InvalidUserID(t *testing.T) {
	h := handlers.NewTransactionHandlerWithModel(&MockTransactionModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/transactions?user_id=abc", nil)
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetTransactions_DBError(t *testing.T) {
	mock := &MockTransactionModel{
		GetAllFn: func(_ context.Context, _ int32) ([]db.ListTransactionsRow, error) { return nil, errors.New("db error") },
	}
	h := handlers.NewTransactionHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/transactions?user_id=1", nil)
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── GetTransaction ────────────────────────────────────────────────────────────

func TestGetTransaction_Success(t *testing.T) {
	mock := &MockTransactionModel{
		GetFn: func(_ context.Context, id, _ int32) (*db.GetTransactionRow, error) {
			return &db.GetTransactionRow{ID: id, UserID: 1}, nil
		},
	}
	h := handlers.NewTransactionHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/transactions/1?user_id=1", nil)
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetTransaction_InvalidID(t *testing.T) {
	h := handlers.NewTransactionHandlerWithModel(&MockTransactionModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/transactions/abc?user_id=1", nil)
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetTransaction_MissingUserID(t *testing.T) {
	h := handlers.NewTransactionHandlerWithModel(&MockTransactionModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/transactions/1", nil)
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ── UpdateTransaction ─────────────────────────────────────────────────────────

func TestUpdateTransaction_Success(t *testing.T) {
	mock := &MockTransactionModel{
		UpdateFn: func(_ context.Context, id, _ int32, _ string, _ float64, _ string, _ *int32, _ *int32, _ *int32, _ pgtype.Date) (*models.Transaction, error) {
			return stubTransaction(id), nil
		},
	}
	h := handlers.NewTransactionHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]any{
		"user_id": 1, "type": "income", "amount": 100.0, "date": "2026-03-01",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/transactions/1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestUpdateTransaction_InvalidID(t *testing.T) {
	h := handlers.NewTransactionHandlerWithModel(&MockTransactionModel{})
	body, _ := json.Marshal(map[string]any{"user_id": 1, "type": "income", "amount": 100.0})
	req := httptest.NewRequest(http.MethodPut, "/api/transactions/abc", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateTransaction_MissingUserID(t *testing.T) {
	h := handlers.NewTransactionHandlerWithModel(&MockTransactionModel{})
	body, _ := json.Marshal(map[string]any{"type": "income", "amount": 100.0})
	req := httptest.NewRequest(http.MethodPut, "/api/transactions/1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateTransaction_InvalidDate(t *testing.T) {
	h := handlers.NewTransactionHandlerWithModel(&MockTransactionModel{})
	body, _ := json.Marshal(map[string]any{"user_id": 1, "type": "income", "amount": 100.0, "date": "bad-date"})
	req := httptest.NewRequest(http.MethodPut, "/api/transactions/1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ── DeleteTransaction ─────────────────────────────────────────────────────────

func TestDeleteTransaction_Success(t *testing.T) {
	mock := &MockTransactionModel{
		DeleteFn: func(_ context.Context, _, _ int32) error { return nil },
	}
	h := handlers.NewTransactionHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodDelete, "/api/transactions/1?user_id=1", nil)
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestDeleteTransaction_InvalidID(t *testing.T) {
	h := handlers.NewTransactionHandlerWithModel(&MockTransactionModel{})
	req := httptest.NewRequest(http.MethodDelete, "/api/transactions/abc?user_id=1", nil)
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDeleteTransaction_MissingUserID(t *testing.T) {
	h := handlers.NewTransactionHandlerWithModel(&MockTransactionModel{})
	req := httptest.NewRequest(http.MethodDelete, "/api/transactions/1", nil)
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDeleteTransaction_DBError(t *testing.T) {
	mock := &MockTransactionModel{
		DeleteFn: func(_ context.Context, _, _ int32) error { return errors.New("db error") },
	}
	h := handlers.NewTransactionHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodDelete, "/api/transactions/1?user_id=1", nil)
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── GetDashboardStats ─────────────────────────────────────────────────────────

func TestGetDashboardStats_Success(t *testing.T) {
	mock := &MockTransactionModel{
		GetDashboardStatsFn: func(_ context.Context, _ int32) (*models.DashboardStats, error) {
			return &models.DashboardStats{TotalIncome: 1000, TotalExpense: 400, TotalBalance: 600}, nil
		},
	}
	h := handlers.NewTransactionHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/stats?user_id=1", nil)
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var stats models.DashboardStats
	json.NewDecoder(w.Body).Decode(&stats)
	if stats.TotalBalance != 600 {
		t.Errorf("expected balance 600, got %f", stats.TotalBalance)
	}
}

func TestGetDashboardStats_MissingUserID(t *testing.T) {
	h := handlers.NewTransactionHandlerWithModel(&MockTransactionModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/stats", nil)
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetDashboardStats_DBError(t *testing.T) {
	mock := &MockTransactionModel{
		GetDashboardStatsFn: func(_ context.Context, _ int32) (*models.DashboardStats, error) {
			return nil, errors.New("db error")
		},
	}
	h := handlers.NewTransactionHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/stats?user_id=1", nil)
	w := httptest.NewRecorder()
	newTransactionRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
