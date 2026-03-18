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
	return &models.Wallet{ID: id, UserID: 1, Name: name, Type: "general"}
}

// ── GetWallets ────────────────────────────────────────────────────────────────

func TestGetWallets_Success(t *testing.T) {
	mock := &MockWalletModel{
		GetAllFn: func(_ context.Context, _ int32) ([]models.Wallet, error) {
			return []models.Wallet{*stubWallet(1, "Cash"), *stubWallet(2, "Bank")}, nil
		},
	}
	h := handlers.NewWalletHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/wallets?user_id=1", nil)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var wallets []models.Wallet
	json.NewDecoder(w.Body).Decode(&wallets)
	if len(wallets) != 2 {
		t.Errorf("expected 2 wallets, got %d", len(wallets))
	}
}

func TestGetWallets_MissingUserID(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/wallets", nil)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetWallets_InvalidUserID(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/wallets?user_id=abc", nil)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetWallets_DBError(t *testing.T) {
	mock := &MockWalletModel{
		GetAllFn: func(_ context.Context, _ int32) ([]models.Wallet, error) { return nil, errors.New("db error") },
	}
	h := handlers.NewWalletHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/wallets?user_id=1", nil)
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
	req := httptest.NewRequest(http.MethodGet, "/api/wallets/1?user_id=1", nil)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetWallet_InvalidID(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/wallets/abc?user_id=1", nil)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetWallet_MissingUserID(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/wallets/1", nil)
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
	req := httptest.NewRequest(http.MethodGet, "/api/wallets/1?user_id=1", nil)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── CreateWallet ──────────────────────────────────────────────────────────────

func TestCreateWallet_Success(t *testing.T) {
	mock := &MockWalletModel{
		CreateFn: func(_ context.Context, userID int32, name, wType string) (*models.Wallet, error) {
			return stubWallet(1, name), nil
		},
	}
	h := handlers.NewWalletHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]any{"user_id": 1, "name": "Cash", "type": "general"})
	req := httptest.NewRequest(http.MethodPost, "/api/wallets", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestCreateWallet_DefaultType(t *testing.T) {
	mock := &MockWalletModel{
		CreateFn: func(_ context.Context, _ int32, name, wType string) (*models.Wallet, error) {
			if wType != "general" {
				t.Errorf("expected default type 'general', got '%s'", wType)
			}
			return stubWallet(1, name), nil
		},
	}
	h := handlers.NewWalletHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]any{"user_id": 1, "name": "Cash"})
	req := httptest.NewRequest(http.MethodPost, "/api/wallets", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestCreateWallet_MissingUserID(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	body, _ := json.Marshal(map[string]any{"name": "Cash"})
	req := httptest.NewRequest(http.MethodPost, "/api/wallets", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateWallet_MissingName(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	body, _ := json.Marshal(map[string]any{"user_id": 1})
	req := httptest.NewRequest(http.MethodPost, "/api/wallets", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateWallet_InvalidJSON(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	req := httptest.NewRequest(http.MethodPost, "/api/wallets", bytes.NewReader([]byte("bad")))
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateWallet_DBError(t *testing.T) {
	mock := &MockWalletModel{
		CreateFn: func(_ context.Context, _ int32, _, _ string) (*models.Wallet, error) { return nil, errors.New("db error") },
	}
	h := handlers.NewWalletHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]any{"user_id": 1, "name": "Cash"})
	req := httptest.NewRequest(http.MethodPost, "/api/wallets", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── UpdateWallet ──────────────────────────────────────────────────────────────

func TestUpdateWallet_Success(t *testing.T) {
	mock := &MockWalletModel{
		UpdateFn: func(_ context.Context, id, _ int32, name, wType string) (*models.Wallet, error) {
			return stubWallet(id, name), nil
		},
	}
	h := handlers.NewWalletHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]any{"user_id": 1, "name": "Savings", "type": "savings"})
	req := httptest.NewRequest(http.MethodPut, "/api/wallets/1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestUpdateWallet_InvalidID(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	body, _ := json.Marshal(map[string]any{"user_id": 1, "name": "x"})
	req := httptest.NewRequest(http.MethodPut, "/api/wallets/abc", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateWallet_MissingUserID(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	body, _ := json.Marshal(map[string]any{"name": "x"})
	req := httptest.NewRequest(http.MethodPut, "/api/wallets/1", bytes.NewReader(body))
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
	req := httptest.NewRequest(http.MethodDelete, "/api/wallets/1?user_id=1", nil)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestDeleteWallet_InvalidID(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	req := httptest.NewRequest(http.MethodDelete, "/api/wallets/abc?user_id=1", nil)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDeleteWallet_MissingUserID(t *testing.T) {
	h := handlers.NewWalletHandlerWithModel(&MockWalletModel{})
	req := httptest.NewRequest(http.MethodDelete, "/api/wallets/1", nil)
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
	req := httptest.NewRequest(http.MethodDelete, "/api/wallets/1?user_id=1", nil)
	w := httptest.NewRecorder()
	newWalletRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
