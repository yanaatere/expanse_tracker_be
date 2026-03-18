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

// newUserRouter wires a handler to a mux route for path-param tests.
func newUserRouter(h *handlers.UserHandler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/api/users", h.GetUsers).Methods(http.MethodGet)
	r.HandleFunc("/api/users", h.CreateUser).Methods(http.MethodPost)
	r.HandleFunc("/api/users/{id}", h.GetUser).Methods(http.MethodGet)
	r.HandleFunc("/api/users/{id}", h.UpdateUser).Methods(http.MethodPut)
	r.HandleFunc("/api/users/{id}", h.DeleteUser).Methods(http.MethodDelete)
	return r
}

// ── GetUsers ─────────────────────────────────────────────────────────────────

func TestGetUsers_Success(t *testing.T) {
	mock := &MockUserModel{
		GetAllFn: func(_ context.Context) ([]models.User, error) {
			return []models.User{
				{ID: 1, Username: "alice", Email: "alice@example.com"},
				{ID: 2, Username: "bob", Email: "bob@example.com"},
			}, nil
		},
	}
	h := handlers.NewUserHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	newUserRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var users []models.User
	json.NewDecoder(w.Body).Decode(&users)
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestGetUsers_DBError(t *testing.T) {
	mock := &MockUserModel{
		GetAllFn: func(_ context.Context) ([]models.User, error) { return nil, errors.New("db error") },
	}
	h := handlers.NewUserHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	newUserRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── GetUser ───────────────────────────────────────────────────────────────────

func TestGetUser_Success(t *testing.T) {
	mock := &MockUserModel{
		GetFn: func(_ context.Context, id int32) (*models.User, error) {
			return &models.User{ID: id, Username: "alice", Email: "alice@example.com"}, nil
		},
	}
	h := handlers.NewUserHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
	w := httptest.NewRecorder()
	newUserRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetUser_InvalidID(t *testing.T) {
	h := handlers.NewUserHandlerWithModel(&MockUserModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/users/abc", nil)
	w := httptest.NewRecorder()
	newUserRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	mock := &MockUserModel{
		GetFn: func(_ context.Context, _ int32) (*models.User, error) { return nil, nil },
	}
	h := handlers.NewUserHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/users/99", nil)
	w := httptest.NewRecorder()
	newUserRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetUser_DBError(t *testing.T) {
	mock := &MockUserModel{
		GetFn: func(_ context.Context, _ int32) (*models.User, error) { return nil, errors.New("db error") },
	}
	h := handlers.NewUserHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
	w := httptest.NewRecorder()
	newUserRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── CreateUser ────────────────────────────────────────────────────────────────

func TestCreateUser_Success(t *testing.T) {
	mock := &MockUserModel{
		CreateFn: func(_ context.Context, username, email string) (*models.User, error) {
			return &models.User{ID: 1, Username: username, Email: email}, nil
		},
	}
	h := handlers.NewUserHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]string{"username": "bob", "email": "bob@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newUserRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestCreateUser_InvalidJSON(t *testing.T) {
	h := handlers.NewUserHandlerWithModel(&MockUserModel{})
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader([]byte("bad")))
	w := httptest.NewRecorder()
	newUserRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateUser_DBError(t *testing.T) {
	mock := &MockUserModel{
		CreateFn: func(_ context.Context, _, _ string) (*models.User, error) { return nil, errors.New("db error") },
	}
	h := handlers.NewUserHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]string{"username": "bob", "email": "bob@example.com"})
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newUserRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── UpdateUser ────────────────────────────────────────────────────────────────

func TestUpdateUser_Success(t *testing.T) {
	mock := &MockUserModel{
		UpdateFn: func(_ context.Context, id int32, username, email string) (*models.User, error) {
			return &models.User{ID: id, Username: username, Email: email}, nil
		},
	}
	h := handlers.NewUserHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]string{"username": "alice2", "email": "alice2@example.com"})
	req := httptest.NewRequest(http.MethodPut, "/api/users/1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newUserRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestUpdateUser_InvalidID(t *testing.T) {
	h := handlers.NewUserHandlerWithModel(&MockUserModel{})
	body, _ := json.Marshal(map[string]string{"username": "x", "email": "x@x.com"})
	req := httptest.NewRequest(http.MethodPut, "/api/users/abc", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newUserRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	mock := &MockUserModel{
		UpdateFn: func(_ context.Context, _ int32, _, _ string) (*models.User, error) { return nil, nil },
	}
	h := handlers.NewUserHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]string{"username": "x", "email": "x@x.com"})
	req := httptest.NewRequest(http.MethodPut, "/api/users/99", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newUserRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// ── DeleteUser ────────────────────────────────────────────────────────────────

func TestDeleteUser_Success(t *testing.T) {
	mock := &MockUserModel{
		DeleteFn: func(_ context.Context, _ int32) error { return nil },
	}
	h := handlers.NewUserHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodDelete, "/api/users/1", nil)
	w := httptest.NewRecorder()
	newUserRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestDeleteUser_InvalidID(t *testing.T) {
	h := handlers.NewUserHandlerWithModel(&MockUserModel{})
	req := httptest.NewRequest(http.MethodDelete, "/api/users/abc", nil)
	w := httptest.NewRecorder()
	newUserRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDeleteUser_DBError(t *testing.T) {
	mock := &MockUserModel{
		DeleteFn: func(_ context.Context, _ int32) error { return errors.New("db error") },
	}
	h := handlers.NewUserHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodDelete, "/api/users/1", nil)
	w := httptest.NewRecorder()
	newUserRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
