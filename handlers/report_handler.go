package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/models"
)

type ReportHandler struct {
	model *models.ReportModel
}

func NewReportHandler(pool *pgxpool.Pool) *ReportHandler {
	return &ReportHandler{model: models.NewReportModel(pool)}
}

// GetReportTransactions godoc
// @Summary Get transactions for report (premium)
// @Description Returns all transactions for the authenticated premium user within a given period.
// @Tags Reports
// @Produce json
// @Param mode  query string true  "Period mode: monthly or annually"
// @Param year  query int    true  "Year (e.g. 2024)"
// @Param month query int    false "Month 1-12, required when mode=monthly"
// @Success 200 {array}  object
// @Failure 400 {object} object
// @Failure 403 {object} object
// @Failure 500 {object} object
// @Router /api/reports/transactions [get]
func (h *ReportHandler) GetReportTransactions(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == 0 {
		WriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	q := r.URL.Query()
	mode := q.Get("mode")
	if mode != "monthly" && mode != "annually" {
		WriteError(w, http.StatusBadRequest, "mode must be 'monthly' or 'annually'")
		return
	}

	yearStr := q.Get("year")
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2000 {
		WriteError(w, http.StatusBadRequest, "year must be a valid integer (e.g. 2024)")
		return
	}

	var start, end time.Time
	if mode == "monthly" {
		monthStr := q.Get("month")
		month, err := strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			WriteError(w, http.StatusBadRequest, "month must be 1-12 when mode=monthly")
			return
		}
		start = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		end = start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	} else {
		start = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		end = time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	}

	startDate := pgtype.Date{Time: start, Valid: true}
	endDate := pgtype.Date{Time: end, Valid: true}

	rows, err := h.model.GetByDateRange(r.Context(), userID, startDate, endDate)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	WriteSuccess(w, http.StatusOK, rows)
}
