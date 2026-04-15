package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/internal/db"
	"github.com/yanaatere/expense_tracking/models"
)

var googleHTTPClient = &http.Client{Timeout: 5 * time.Second}

type AuthHandler struct {
	userModel UserModelInterface
}

func NewAuthHandler(database db.DBTX) *AuthHandler {
	return &AuthHandler{
		userModel: models.NewUserModel(database),
	}
}

func NewAuthHandlerWithModel(model UserModelInterface) *AuthHandler {
	return &AuthHandler{userModel: model}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type AuthResponse struct {
	ID        int32  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Token     string `json:"token"`
	IsPremium bool   `json:"is_premium"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

// Register handler
// @Summary Register a new user
// @Description Register a new user and return JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register request"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} MessageResponse
// @Failure 409 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		WriteError(w, http.StatusBadRequest, "Username, email, and password are required")
		return
	}

	// Check if user already exists
	existingUser, _ := h.userModel.GetByEmail(r.Context(), req.Email)
	if existingUser != nil {
		WriteError(w, http.StatusConflict, "Email already registered")
		return
	}

	// Check if username already exists
	existingUser, _ = h.userModel.GetByUsername(r.Context(), req.Username)
	if existingUser != nil {
		WriteError(w, http.StatusConflict, "Username already taken")
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Error creating account")
		return
	}

	// Create user
	user, err := h.userModel.CreateWithPassword(r.Context(), req.Username, req.Email, hashedPassword)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Error creating account")
		return
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	WriteSuccess(w, http.StatusCreated, AuthResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Token:     token,
		IsPremium: user.IsPremium,
	})
}

// Login handler
// @Summary Login user
// @Description Authenticate a user and return JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} MessageResponse
// @Failure 401 {object} MessageResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		WriteError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	// Get user by email
	user, err := h.userModel.GetByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		WriteError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Compare password
	if !auth.ComparePassword(user.Password, req.Password) {
		WriteError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	WriteSuccess(w, http.StatusOK, AuthResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Token:     token,
		IsPremium: user.IsPremium,
	})
}

// ForgotPassword handler
// @Summary Forgot password
// @Description Request password reset token by email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "Forgot password request"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} MessageResponse
// @Router /api/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		WriteError(w, http.StatusBadRequest, "Email is required")
		return
	}

	// Get user by email
	user, err := h.userModel.GetByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		// Don't reveal whether email exists
		WriteSuccess(w, http.StatusOK, MessageResponse{
			Message: "If email exists in our system, a password reset link will be sent",
		})
		return
	}

	// Generate reset token
	resetToken, err := auth.GenerateResetToken()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Error generating reset token")
		return
	}

	// Set reset token with expiration (1 hour)
	expiresAt := time.Now().Add(1 * time.Hour)
	_, err = h.userModel.SetPasswordResetToken(r.Context(), user.ID, resetToken, expiresAt)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Error processing request")
		return
	}

	// Send email (placeholder)
	err = auth.SendPasswordResetEmail(user.Email, resetToken)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Error sending email")
		return
	}

	WriteSuccess(w, http.StatusOK, MessageResponse{
		Message: "If email exists in our system, a password reset link will be sent",
	})
}

// ResetPassword handler
// @Summary Reset password
// @Description Reset password using reset token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ResetPasswordRequest true "Reset password request"
// @Success 200 {object} MessageResponse
// @Failure 400 {object} MessageResponse
// @Failure 401 {object} MessageResponse
// @Router /api/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		WriteError(w, http.StatusBadRequest, "Token and new password are required")
		return
	}

	// Get user by reset token
	user, err := h.userModel.GetByResetToken(r.Context(), req.Token)
	if err != nil || user == nil {
		WriteError(w, http.StatusUnauthorized, "Invalid or expired reset token")
		return
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Error resetting password")
		return
	}

	// Update password and clear reset token
	updatedUser, err := h.userModel.UpdatePassword(r.Context(), user.ID, hashedPassword)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Error updating password")
		return
	}

	// Clear the reset token
	_, err = h.userModel.ClearPasswordResetToken(r.Context(), updatedUser.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Error clearing reset token")
		return
	}

	WriteSuccess(w, http.StatusOK, MessageResponse{
		Message: "Password has been reset successfully",
	})
}

type GoogleLoginRequest struct {
	IDToken string `json:"id_token"`
}

// GoogleLogin verifies a Google ID token and returns a Monex JWT.
// It creates a new account automatically if the email is not yet registered.
// @Summary Google Sign-In
// @Description Verify a Google ID token and return a Monex JWT
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body GoogleLoginRequest true "Google ID token"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} MessageResponse
// @Failure 401 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/auth/google [post]
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	var req GoogleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if req.IDToken == "" {
		WriteError(w, http.StatusBadRequest, "id_token is required")
		return
	}

	// Verify token via Google's tokeninfo endpoint (no extra dependency needed).
	resp, err := googleHTTPClient.Get(
		"https://oauth2.googleapis.com/tokeninfo?id_token=" + req.IDToken,
	)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "Failed to verify Google token")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		WriteError(w, http.StatusUnauthorized, "Invalid Google token")
		return
	}

	var info map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to parse token info")
		return
	}

	// Optional audience check — skip if GOOGLE_CLIENT_ID is not set.
	if clientID := os.Getenv("GOOGLE_CLIENT_ID"); clientID != "" {
		if info["aud"] != clientID {
			WriteError(w, http.StatusUnauthorized, "Token audience mismatch")
			return
		}
	}

	email := info["email"]
	if email == "" {
		WriteError(w, http.StatusUnauthorized, "Google account has no email")
		return
	}
	if info["email_verified"] != "true" {
		WriteError(w, http.StatusUnauthorized, "Google email is not verified")
		return
	}

	// Find existing user or create a new one.
	user, err := h.userModel.GetByEmail(r.Context(), email)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			WriteError(w, http.StatusInternalServerError, "Error looking up account")
			return
		}
		// New user — derive username from email prefix.
		username := googleDeriveUsername(email, info["name"])
		user, err = h.userModel.CreateWithPassword(r.Context(), username, email, "")
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "Error creating account")
			return
		}
	}

	token, err := auth.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Error generating token")
		return
	}

	WriteSuccess(w, http.StatusOK, AuthResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Token:     token,
		IsPremium: user.IsPremium,
	})
}

// GetMe godoc
// @Summary Get current user profile
// @Description Returns the authenticated user's profile including premium status
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} MessageResponse
// @Router /api/auth/me [get]
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	user, err := h.userModel.Get(r.Context(), userID)
	if err != nil || user == nil {
		WriteError(w, http.StatusNotFound, "User not found")
		return
	}
	WriteSuccess(w, http.StatusOK, map[string]any{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"is_premium": user.IsPremium,
	})
}

// SetPremium godoc
// @Summary Set premium status for current user
// @Description Update the authenticated user's premium status
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body map[string]bool true "Premium status"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /api/users/{id}/premium [put]
func (h *AuthHandler) SetPremium(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	var body struct {
		IsPremium bool `json:"is_premium"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	user, err := h.userModel.SetPremium(r.Context(), userID, body.IsPremium)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to update premium status")
		return
	}
	WriteSuccess(w, http.StatusOK, map[string]any{"is_premium": user.IsPremium})
}

// googleDeriveUsername turns an email/name into a safe lowercase username.
func googleDeriveUsername(email, name string) string {
	if idx := strings.Index(email, "@"); idx > 0 {
		base := strings.ToLower(email[:idx])
		// Replace dots/hyphens/plus with underscore.
		replacer := strings.NewReplacer(".", "_", "-", "_", "+", "_")
		return replacer.Replace(base)
	}
	if name != "" {
		return strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	}
	return "user"
}
