package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/handlers"
	"github.com/yanaatere/expense_tracking/models"
)

func newWalletRouter(h *handlers.WalletHandler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/api/wallets", h.GetWallets).Methods(http.MethodGet)
	r.HandleFunc("/api/wallets", h.CreateWallet).Methods(http.MethodPost)
	r.HandleFunc("/api/wallets/{id}", h.GetWallet).Methods(http.MethodGet)
	r.HandleFunc("/api/wallets/{id}", h.UpdateWallet).Methods(http.MethodPut)
	r.HandleFunc("/api/wallets/{id}", h.DeleteWallet).Methods(http.MethodDelete)
	return r
}

func stubWallet(id int32, name string) *models.Wallet {
	return &models.Wallet{ID: id, UserID: 1, Name: name, Type: "bank", Currency: "IDR"}
}

// withUserID injects a user ID into the request context, simulating JWT middleware.
func withUserID(req *http.Request, userID int32) *http.Request {
	ctx := context.WithValue(req.Context(), auth.UserContextKey, userID)
	return req.WithContext(ctx)
}

// ── GetWallets ────────────────────────────────────────────────────────────────

func TestGetWallets_Success(t *testing.T) {
	mock := &MockWalletModel{
		GetAllFn: func(_ context.Context, _ int32) ([]models.Wallet, error) {
			return []models.Wallet{*stubWallet(1, "Cash"), *stubWallet(2, "Bank")}, nil
		},
	}
	h := handlers.NewWalletHandlerWithModel(mock)
	req := withUserID(httptest.NewRequest(http.MethodGet, "/api/wallets", nil), 1)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetWallets_Unauthorized(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/wallets", nil)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestGetWallets_DBError(t *testing.T) {
	mock := &MockWalletModel{
		GetAllFn: func(_ context.Context, _ int32) ([]models.Wallet, error) { return nil, errors.New("db error") },
	}
	h := handlers.NewWalletHandlerWithModel(mock)
	req := withUserID(httptest.NewRequest(http.MethodGet, "/api/wallets", nil), 1)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── GetWallet ─────────────────────────────────────────────────────────────────

func TestGetWallet_Success(t *testing.T) {
	mock := &MockWalletModel{
		GetFn: func(_ context.Context, id, _ int32) (*models.Wallet, error) { return stubWallet(id, "Cash"), nil },
	}
	h := handlers.NewWalletHandlerWithModel(mock)
	req := withUserID(httptest.NewRequest(http.MethodGet, "/api/wallets/1", nil), 1)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetWallet_Unauthorized(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/wallets/1", nil)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestGetWallet_InvalidID(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	req := withUserID(httptest.NewRequest(http.MethodGet, "/api/wallets/abc", nil), 1)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetWallet_DBError(t *testing.T) {
	mock := &MockWalletModel{
		GetFn: func(_ context.Context, _, _ int32) (*models.Wallet, error) { return nil, errors.New("db error") },
	}
	h := handlers.NewWalletHandlerWithModel(mock)
	req := withUserID(httptest.NewRequest(http.MethodGet, "/api/wallets/1", nil), 1)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── CreateWallet ──────────────────────────────────────────────────────────────

func TestCreateWallet_Success(t *testing.T) {
	mock := &MockWalletModel{
		CreateFn: func(_ context.Context, userID int32, name, wType, currency string, balance float64, goals *string, backdropImage *string) (*models.Wallet, error) {
			return stubWallet(1, name), nil
		},
	}
	h := handlers.NewWalletHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]any{"name": "Cash", "type": "bank", "currency": "IDR", "balance": 100000})
	req := withUserID(httptest.NewRequest(http.MethodPost, "/api/wallets", bytes.NewReader(body)), 1)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestCreateWallet_Unauthorized(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	body, _ := json.Marshal(map[string]any{"name": "Cash", "type": "bank", "currency": "IDR"})
	req := httptest.NewRequest(http.MethodPost, "/api/wallets", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestCreateWallet_MissingName(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	body, _ := json.Marshal(map[string]any{"type": "bank", "currency": "IDR"})
	req := withUserID(httptest.NewRequest(http.MethodPost, "/api/wallets", bytes.NewReader(body)), 1)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateWallet_MissingCurrency(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	body, _ := json.Marshal(map[string]any{"name": "Cash", "type": "bank"})
	req := withUserID(httptest.NewRequest(http.MethodPost, "/api/wallets", bytes.NewReader(body)), 1)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateWallet_MissingType(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	body, _ := json.Marshal(map[string]any{"name": "Cash", "currency": "IDR"})
	req := withUserID(httptest.NewRequest(http.MethodPost, "/api/wallets", bytes.NewReader(body)), 1)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateWallet_InvalidJSON(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	req := withUserID(httptest.NewRequest(http.MethodPost, "/api/wallets", bytes.NewReader([]byte("bad"))), 1)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateWallet_DBError(t *testing.T) {
	mock := &MockWalletModel{
		CreateFn: func(_ context.Context, _ int32, _, _, _ string, _ float64, _ *string, _ *string) (*models.Wallet, error) {
			return nil, errors.New("db error")
		},
	}
	h := handlers.NewWalletHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]any{"name": "Cash", "type": "bank", "currency": "IDR"})
	req := withUserID(httptest.NewRequest(http.MethodPost, "/api/wallets", bytes.NewReader(body)), 1)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── UpdateWallet ──────────────────────────────────────────────────────────────

func TestUpdateWallet_Success(t *testing.T) {
	mock := &MockWalletModel{
		UpdateFn: func(_ context.Context, id, _ int32, name, wType, currency string, balance float64, goals *string, backdropImage *string) (*models.Wallet, error) {
			return stubWallet(id, name), nil
		},
	}
	h := handlers.NewWalletHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]any{"name": "Savings", "type": "bank", "currency": "IDR", "balance": 500000})
	req := withUserID(httptest.NewRequest(http.MethodPut, "/api/wallets/1", bytes.NewReader(body)), 1)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestUpdateWallet_Unauthorized(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	body, _ := json.Marshal(map[string]any{"name": "x", "type": "bank", "currency": "IDR"})
	req := httptest.NewRequest(http.MethodPut, "/api/wallets/1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestUpdateWallet_InvalidID(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	body, _ := json.Marshal(map[string]any{"name": "x", "type": "bank", "currency": "IDR"})
	req := withUserID(httptest.NewRequest(http.MethodPut, "/api/wallets/abc", bytes.NewReader(body)), 1)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ── DeleteWallet ──────────────────────────────────────────────────────────────

func TestDeleteWallet_Success(t *testing.T) {
	mock := &MockWalletModel{
		DeleteFn: func(_ context.Context, _, _ int32) error { return nil },
	}
	h := handlers.NewWalletHandlerWithModel(mock)
	req := withUserID(httptest.NewRequest(http.MethodDelete, "/api/wallets/1", nil), 1)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestDeleteWallet_Unauthorized(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	req := httptest.NewRequest(http.MethodDelete, "/api/wallets/1", nil)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestDeleteWallet_InvalidID(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	req := withUserID(httptest.NewRequest(http.MethodDelete, "/api/wallets/abc", nil), 1)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDeleteWallet_DBError(t *testing.T) {
	mock := &MockWalletModel{
		DeleteFn: func(_ context.Context, _, _ int32) error { return errors.New("db error") },
	}
	h := handlers.NewWalletHandlerWithModel(mock)
	req := withUserID(httptest.NewRequest(http.MethodDelete, "/api/wallets/1", nil), 1)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
