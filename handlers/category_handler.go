package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/yanaatere/expense_tracking/internal/db"
	"github.com/yanaatere/expense_tracking/models"
)

type CategoryHandler struct {
	model CategoryModelInterface
}

func NewCategoryHandler(db db.DBTX) *CategoryHandler {
	return &CategoryHandler{
		model: models.NewCategoryModel(db),
	}
}

func NewCategoryHandlerWithModel(model CategoryModelInterface) *CategoryHandler {
	return &CategoryHandler{model: model}
}

type CategoryInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    *int   `json:"parent_id"`
}

// @Summary List categories
// @Description Get top-level categories (parent_id IS NULL) (protected)
// @Tags Categories
// @Produce json
// @Success 200 {array} object
// @Failure 500 {object} MessageResponse
// @Router /api/categories [get]
func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.model.GetAll(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, categories)
}

// @Summary Create category
// @Description Create a new category or subcategory (protected)
// @Tags Categories
// @Accept json
// @Produce json
// @Param request body CategoryInput true "Category create request"
// @Success 201 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/categories [post]
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var input CategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if input.Name == "" {
		WriteError(w, http.StatusBadRequest, "Name is required")
		return
	}

	var parentID *int32
	if input.ParentID != nil {
		id := int32(*input.ParentID)
		parentID = &id
	}
	category, err := h.model.Create(r.Context(), input.Name, input.Description, parentID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusCreated, category)
}

// @Summary Get category
// @Description Get single category by id (protected)
// @Tags Categories
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} object
// @Failure 400 {object} MessageResponse
// @Failure 404 {object} MessageResponse
// @Router /api/categories/{id} [get]
func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	category, err := h.model.Get(r.Context(), int32(id))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if category == nil {
		WriteError(w, http.StatusNotFound, "Category not found")
		return
	}

	WriteSuccess(w, http.StatusOK, category)
}

// @Summary Update category
// @Description Update category metadata (protected)
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param request body CategoryInput true "Category update request"
// @Success 200 {object} object
// @Failure 400 {object} MessageResponse
// @Router /api/categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	var input CategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	category, err := h.model.Update(r.Context(), int32(id), input.Name, input.Description)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, category)
}

// @Summary List sub-categories
// @Description Get subcategories of a category (protected)
// @Tags Categories
// @Produce json
// @Param id path int true "Parent Category ID"
// @Success 200 {array} object
// @Failure 400 {object} MessageResponse
// @Router /api/categories/{id}/sub-categories [get]
func (h *CategoryHandler) GetSubCategories(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	subCategories, err := h.model.GetSubCategories(r.Context(), int32(id))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, subCategories)
}

// @Summary Delete category
// @Description Delete category by id (protected)
// @Tags Categories
// @Param id path int true "Category ID"
// @Success 204 {object} MessageResponse
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid category ID")
		return
	}

	err = h.model.Delete(r.Context(), int32(id))
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
