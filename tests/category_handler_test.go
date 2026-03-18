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

func newCategoryRouter(h *handlers.CategoryHandler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/api/categories", h.GetCategories).Methods(http.MethodGet)
	r.HandleFunc("/api/categories", h.CreateCategory).Methods(http.MethodPost)
	r.HandleFunc("/api/categories/{id}", h.GetCategory).Methods(http.MethodGet)
	r.HandleFunc("/api/categories/{id}", h.UpdateCategory).Methods(http.MethodPut)
	r.HandleFunc("/api/categories/{id}", h.DeleteCategory).Methods(http.MethodDelete)
	r.HandleFunc("/api/categories/{id}/sub-categories", h.GetSubCategories).Methods(http.MethodGet)
	return r
}

func stubCategory(id int32, name string) *models.Category {
	return &models.Category{ID: id, Name: name}
}

// ── GetCategories ─────────────────────────────────────────────────────────────

func TestGetCategories_Success(t *testing.T) {
	mock := &MockCategoryModel{
		GetAllFn: func(_ context.Context) ([]models.Category, error) {
			return []models.Category{*stubCategory(1, "Food"), *stubCategory(2, "Transport")}, nil
		},
	}
	h := handlers.NewCategoryHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var cats []models.Category
	json.NewDecoder(w.Body).Decode(&cats)
	if len(cats) != 2 {
		t.Errorf("expected 2 categories, got %d", len(cats))
	}
}

func TestGetCategories_DBError(t *testing.T) {
	mock := &MockCategoryModel{
		GetAllFn: func(_ context.Context) ([]models.Category, error) { return nil, errors.New("db error") },
	}
	h := handlers.NewCategoryHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── GetCategory ───────────────────────────────────────────────────────────────

func TestGetCategory_Success(t *testing.T) {
	mock := &MockCategoryModel{
		GetFn: func(_ context.Context, id int32) (*models.Category, error) { return stubCategory(id, "Food"), nil },
	}
	h := handlers.NewCategoryHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/categories/1", nil)
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetCategory_InvalidID(t *testing.T) {
	h := handlers.NewCategoryHandlerWithModel(&MockCategoryModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/categories/abc", nil)
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetCategory_NotFound(t *testing.T) {
	mock := &MockCategoryModel{
		GetFn: func(_ context.Context, _ int32) (*models.Category, error) { return nil, nil },
	}
	h := handlers.NewCategoryHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/categories/99", nil)
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestGetCategory_DBError(t *testing.T) {
	mock := &MockCategoryModel{
		GetFn: func(_ context.Context, _ int32) (*models.Category, error) { return nil, errors.New("db error") },
	}
	h := handlers.NewCategoryHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/categories/1", nil)
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── CreateCategory ────────────────────────────────────────────────────────────

func TestCreateCategory_Success(t *testing.T) {
	mock := &MockCategoryModel{
		CreateFn: func(_ context.Context, name, _ string, _ *int32) (*models.Category, error) {
			return stubCategory(1, name), nil
		},
	}
	h := handlers.NewCategoryHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]string{"name": "Food", "description": "Food expenses"})
	req := httptest.NewRequest(http.MethodPost, "/api/categories", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestCreateCategory_WithParent_Success(t *testing.T) {
	mock := &MockCategoryModel{
		CreateFn: func(_ context.Context, name, _ string, pid *int32) (*models.Category, error) {
			return stubCategory(2, name), nil
		},
	}
	h := handlers.NewCategoryHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]any{"name": "Snacks", "parent_id": 1})
	req := httptest.NewRequest(http.MethodPost, "/api/categories", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}
}

func TestCreateCategory_MissingName(t *testing.T) {
	h := handlers.NewCategoryHandlerWithModel(&MockCategoryModel{})
	body, _ := json.Marshal(map[string]string{"description": "no name"})
	req := httptest.NewRequest(http.MethodPost, "/api/categories", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateCategory_InvalidJSON(t *testing.T) {
	h := handlers.NewCategoryHandlerWithModel(&MockCategoryModel{})
	req := httptest.NewRequest(http.MethodPost, "/api/categories", bytes.NewReader([]byte("bad")))
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateCategory_DBError(t *testing.T) {
	mock := &MockCategoryModel{
		CreateFn: func(_ context.Context, _, _ string, _ *int32) (*models.Category, error) {
			return nil, errors.New("db error")
		},
	}
	h := handlers.NewCategoryHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]string{"name": "Food"})
	req := httptest.NewRequest(http.MethodPost, "/api/categories", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── UpdateCategory ────────────────────────────────────────────────────────────

func TestUpdateCategory_Success(t *testing.T) {
	mock := &MockCategoryModel{
		UpdateFn: func(_ context.Context, id int32, name, _ string) (*models.Category, error) {
			return stubCategory(id, name), nil
		},
	}
	h := handlers.NewCategoryHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]string{"name": "Updated Food"})
	req := httptest.NewRequest(http.MethodPut, "/api/categories/1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestUpdateCategory_InvalidID(t *testing.T) {
	h := handlers.NewCategoryHandlerWithModel(&MockCategoryModel{})
	body, _ := json.Marshal(map[string]string{"name": "x"})
	req := httptest.NewRequest(http.MethodPut, "/api/categories/abc", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateCategory_DBError(t *testing.T) {
	mock := &MockCategoryModel{
		UpdateFn: func(_ context.Context, _ int32, _, _ string) (*models.Category, error) {
			return nil, errors.New("db error")
		},
	}
	h := handlers.NewCategoryHandlerWithModel(mock)
	body, _ := json.Marshal(map[string]string{"name": "x"})
	req := httptest.NewRequest(http.MethodPut, "/api/categories/1", bytes.NewReader(body))
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── GetSubCategories ──────────────────────────────────────────────────────────

func TestGetSubCategories_Success(t *testing.T) {
	mock := &MockCategoryModel{
		GetSubCategoriesFn: func(_ context.Context, _ int32) ([]models.Category, error) {
			return []models.Category{*stubCategory(2, "Snacks"), *stubCategory(3, "Drinks")}, nil
		},
	}
	h := handlers.NewCategoryHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/categories/1/sub-categories", nil)
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var subs []models.Category
	json.NewDecoder(w.Body).Decode(&subs)
	if len(subs) != 2 {
		t.Errorf("expected 2 subcategories, got %d", len(subs))
	}
}

func TestGetSubCategories_InvalidID(t *testing.T) {
	h := handlers.NewCategoryHandlerWithModel(&MockCategoryModel{})
	req := httptest.NewRequest(http.MethodGet, "/api/categories/abc/sub-categories", nil)
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetSubCategories_DBError(t *testing.T) {
	mock := &MockCategoryModel{
		GetSubCategoriesFn: func(_ context.Context, _ int32) ([]models.Category, error) {
			return nil, errors.New("db error")
		},
	}
	h := handlers.NewCategoryHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/categories/1/sub-categories", nil)
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ── DeleteCategory ────────────────────────────────────────────────────────────

func TestDeleteCategory_Success(t *testing.T) {
	mock := &MockCategoryModel{
		DeleteFn: func(_ context.Context, _ int32) error { return nil },
	}
	h := handlers.NewCategoryHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodDelete, "/api/categories/1", nil)
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestDeleteCategory_InvalidID(t *testing.T) {
	h := handlers.NewCategoryHandlerWithModel(&MockCategoryModel{})
	req := httptest.NewRequest(http.MethodDelete, "/api/categories/abc", nil)
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDeleteCategory_DBError(t *testing.T) {
	mock := &MockCategoryModel{
		DeleteFn: func(_ context.Context, _ int32) error { return errors.New("db error") },
	}
	h := handlers.NewCategoryHandlerWithModel(mock)
	req := httptest.NewRequest(http.MethodDelete, "/api/categories/1", nil)
	w := httptest.NewRecorder()
	newCategoryRouter(h).ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
