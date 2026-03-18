package tests

import (
	"context"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/handlers"
	"github.com/yanaatere/expense_tracking/models"
)

// ── helpers ─────────────────────────────────────────────────────────────────

func stubUser(id int32, username, email string) *models.User {
	return &models.User{ID: id, Username: username, Email: email}
}

func postJSON(handler func(http.ResponseWriter, *http.Request), path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	w := httptest.NewRecorder()
	handler(w, req)
	return w
}

// ── Register ─────────────────────────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	mock := &MockUserModel{
		GetByEmailFn:    func(_ context.Context, _ string) (*models.User, error) { return nil, errors.New("not found") },
		GetByUsernameFn: func(_ context.Context, _ string) (*models.User, error) { return nil, errors.New("not found") },
		CreateWithPasswordFn: func(_ context.Context, username, email, _ string) (*models.User, error) {
			return stubUser(1, username, email), nil
		},
	}
	h := handlers.NewAuthHandlerWithModel(mock)

	w := postJSON(h.Register, "/api/auth/register", map[string]string{
		"username": "alice", "email": "alice@example.com", "password": "secret123",
	})

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
	var resp handlers.AuthResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Email != "alice@example.com" {
		t.Errorf("unexpected email: %s", resp.Email)
	}
	if resp.Token == "" {
		t.Error("expected JWT token in response")
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	h := handlers.NewAuthHandlerWithModel(&MockUserModel{})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte("not-json")))
	w := httptest.NewRecorder()
	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegister_MissingUsername(t *testing.T) {
	h := handlers.NewAuthHandlerWithModel(&MockUserModel{})
	w := postJSON(h.Register, "/api/auth/register", map[string]string{"email": "a@b.com", "password": "pass"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegister_MissingEmail(t *testing.T) {
	h := handlers.NewAuthHandlerWithModel(&MockUserModel{})
	w := postJSON(h.Register, "/api/auth/register", map[string]string{"username": "alice", "password": "pass"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegister_MissingPassword(t *testing.T) {
	h := handlers.NewAuthHandlerWithModel(&MockUserModel{})
	w := postJSON(h.Register, "/api/auth/register", map[string]string{"username": "alice", "email": "a@b.com"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	mock := &MockUserModel{
		GetByEmailFn: func(_ context.Context, _ string) (*models.User, error) {
			return stubUser(1, "existing", "existing@example.com"), nil
		},
	}
	h := handlers.NewAuthHandlerWithModel(mock)
	w := postJSON(h.Register, "/api/auth/register", map[string]string{
		"username": "alice", "email": "existing@example.com", "password": "pass",
	})

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestRegister_UsernameAlreadyTaken(t *testing.T) {
	mock := &MockUserModel{
		GetByEmailFn:    func(_ context.Context, _ string) (*models.User, error) { return nil, errors.New("not found") },
		GetByUsernameFn: func(_ context.Context, _ string) (*models.User, error) { return stubUser(1, "alice", "other@x.com"), nil },
	}
	h := handlers.NewAuthHandlerWithModel(mock)
	w := postJSON(h.Register, "/api/auth/register", map[string]string{
		"username": "alice", "email": "new@example.com", "password": "pass",
	})

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

// ── Login ────────────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	hashed, _ := auth.HashPassword("secret123")
	mock := &MockUserModel{
		GetByEmailFn: func(_ context.Context, _ string) (*models.User, error) {
			u := stubUser(1, "alice", "alice@example.com")
			u.Password = hashed
			return u, nil
		},
	}
	h := handlers.NewAuthHandlerWithModel(mock)
	w := postJSON(h.Login, "/api/auth/login", map[string]string{
		"email": "alice@example.com", "password": "secret123",
	})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp handlers.AuthResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Token == "" {
		t.Error("expected JWT token in response")
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	h := handlers.NewAuthHandlerWithModel(&MockUserModel{})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte("{bad}")))
	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLogin_MissingEmail(t *testing.T) {
	h := handlers.NewAuthHandlerWithModel(&MockUserModel{})
	w := postJSON(h.Login, "/api/auth/login", map[string]string{"password": "pass"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLogin_MissingPassword(t *testing.T) {
	h := handlers.NewAuthHandlerWithModel(&MockUserModel{})
	w := postJSON(h.Login, "/api/auth/login", map[string]string{"email": "a@b.com"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	mock := &MockUserModel{
		GetByEmailFn: func(_ context.Context, _ string) (*models.User, error) { return nil, errors.New("not found") },
	}
	h := handlers.NewAuthHandlerWithModel(mock)
	w := postJSON(h.Login, "/api/auth/login", map[string]string{"email": "x@x.com", "password": "pass"})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hashed, _ := auth.HashPassword("correct")
	mock := &MockUserModel{
		GetByEmailFn: func(_ context.Context, _ string) (*models.User, error) {
			u := stubUser(1, "alice", "alice@example.com")
			u.Password = hashed
			return u, nil
		},
	}
	h := handlers.NewAuthHandlerWithModel(mock)
	w := postJSON(h.Login, "/api/auth/login", map[string]string{"email": "alice@example.com", "password": "wrong"})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// ── ForgotPassword ───────────────────────────────────────────────────────────

func TestForgotPassword_MissingEmail(t *testing.T) {
	h := handlers.NewAuthHandlerWithModel(&MockUserModel{})
	w := postJSON(h.ForgotPassword, "/api/auth/forgot-password", map[string]string{})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestForgotPassword_EmailNotFound_DoesNotReveal(t *testing.T) {
	mock := &MockUserModel{
		GetByEmailFn: func(_ context.Context, _ string) (*models.User, error) { return nil, errors.New("not found") },
	}
	h := handlers.NewAuthHandlerWithModel(mock)
	w := postJSON(h.ForgotPassword, "/api/auth/forgot-password", map[string]string{"email": "ghost@example.com"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (non-revealing), got %d", w.Code)
	}
}

func TestForgotPassword_Success(t *testing.T) {
	user := stubUser(1, "alice", "alice@example.com")
	mock := &MockUserModel{
		GetByEmailFn: func(_ context.Context, _ string) (*models.User, error) { return user, nil },
		SetPasswordResetTokenFn: func(_ context.Context, _ int32, _ string, _ time.Time) (*models.User, error) {
			return user, nil
		},
	}
	h := handlers.NewAuthHandlerWithModel(mock)
	w := postJSON(h.ForgotPassword, "/api/auth/forgot-password", map[string]string{"email": "alice@example.com"})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// ── ResetPassword ────────────────────────────────────────────────────────────

func TestResetPassword_MissingToken(t *testing.T) {
	h := handlers.NewAuthHandlerWithModel(&MockUserModel{})
	w := postJSON(h.ResetPassword, "/api/auth/reset-password", map[string]string{"new_password": "newpass"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestResetPassword_MissingNewPassword(t *testing.T) {
	h := handlers.NewAuthHandlerWithModel(&MockUserModel{})
	w := postJSON(h.ResetPassword, "/api/auth/reset-password", map[string]string{"token": "abc123"})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestResetPassword_InvalidToken(t *testing.T) {
	mock := &MockUserModel{
		GetByResetTokenFn: func(_ context.Context, _ string) (*models.User, error) { return nil, errors.New("not found") },
	}
	h := handlers.NewAuthHandlerWithModel(mock)
	w := postJSON(h.ResetPassword, "/api/auth/reset-password", map[string]string{
		"token": "badtoken", "new_password": "newpass",
	})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestResetPassword_Success(t *testing.T) {
	user := stubUser(1, "alice", "alice@example.com")
	mock := &MockUserModel{
		GetByResetTokenFn:         func(_ context.Context, _ string) (*models.User, error) { return user, nil },
		UpdatePasswordFn:          func(_ context.Context, _ int32, _ string) (*models.User, error) { return user, nil },
		ClearPasswordResetTokenFn: func(_ context.Context, _ int32) (*models.User, error) { return user, nil },
	}
	h := handlers.NewAuthHandlerWithModel(mock)
	w := postJSON(h.ResetPassword, "/api/auth/reset-password", map[string]string{
		"token": "validtoken", "new_password": "newpass",
	})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
