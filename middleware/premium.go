package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/internal/db"
)

// PremiumRequired returns a middleware that allows only premium users through.
// Must be chained after auth.JWTMiddleware so that the user ID is in context.
func PremiumRequired(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := auth.GetUserIDFromContext(r.Context())
			if userID == 0 {
				writeForbidden(w, "unauthorized")
				return
			}

			q := db.New(pool)
			user, err := q.GetUser(r.Context(), userID)
			if err != nil || !user.IsPremium {
				writeForbidden(w, "premium subscription required")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeForbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"msgId":  "forbidden",
		"status": "error",
		"data":   map[string]string{"message": message},
	})
}
