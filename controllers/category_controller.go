package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/yanaatere/expense_tracking/internal/db"
	"github.com/yanaatere/expense_tracking/models"
)

type CategoryController struct {
	model *models.CategoryModel
}

func NewCategoryController(db db.DBTX) *CategoryController {
	return &CategoryController{
		model: models.NewCategoryModel(db),
	}
}

type CategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *CategoryController) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/categories", c.GetCategories).Methods("GET")
	router.HandleFunc("/api/categories/{id:[0-9]+}", c.GetCategory).Methods("GET")
	router.HandleFunc("/api/categories", c.CreateCategory).Methods("POST")
	router.HandleFunc("/api/categories/{id:[0-9]+}", c.UpdateCategory).Methods("PUT")
	router.HandleFunc("/api/categories/{id:[0-9]+}", c.DeleteCategory).Methods("DELETE")
}

func (c *CategoryController) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := c.model.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func (c *CategoryController) GetCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	category, err := c.model.Get(r.Context(), int32(id))
	if err != nil {
		// if err == sql.ErrNoRows {
		// 	http.Error(w, "Category not found", http.StatusNotFound)
		// 	return
		// }
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

func (c *CategoryController) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	category, err := c.model.Create(r.Context(), req.Name, req.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

func (c *CategoryController) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var req CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	category, err := c.model.Update(r.Context(), int32(id), req.Name, req.Description)
	if err != nil {
		// if err == sql.ErrNoRows {
		// 	http.Error(w, "Category not found", http.StatusNotFound)
		// 	return
		// }
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

func (c *CategoryController) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	err = c.model.Delete(r.Context(), int32(id))
	if err != nil {
		// if err == sql.ErrNoRows {
		// 	http.Error(w, "Category not found", http.StatusNotFound)
		// 	return
		// }
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
