package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yanaatere/expense_tracking/auth"
	"github.com/yanaatere/expense_tracking/internal/db"
	"github.com/yanaatere/expense_tracking/models"
)

type AuthHandler struct {
	userModel *models.UserModel
}

func NewAuthHandler(database db.DBTX) *AuthHandler {
	return &AuthHandler{
		userModel: models.NewUserModel(database),
	}
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
	ID       int32  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Token    string `json:"token"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

// Register handler
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Username, email, and password are required", http.StatusBadRequest)
		return
	}

	// Check if user already exists
	existingUser, _ := h.userModel.GetByEmail(r.Context(), req.Email)
	if existingUser != nil {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}

	// Check if username already exists
	existingUser, _ = h.userModel.GetByUsername(r.Context(), req.Username)
	if existingUser != nil {
		http.Error(w, "Username already taken", http.StatusConflict)
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Error creating account", http.StatusInternalServerError)
		return
	}

	// Create user
	user, err := h.userModel.CreateWithPassword(r.Context(), req.Username, req.Email, hashedPassword)
	if err != nil {
		http.Error(w, "Error creating account: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(AuthResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Token:    token,
	})
}

// Login handler
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Get user by email
	user, err := h.userModel.GetByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Compare password
	if !auth.ComparePassword(user.Password, req.Password) {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generate token
	token, err := auth.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Token:    token,
	})
}

// ForgotPassword handler
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Get user by email
	user, err := h.userModel.GetByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		// Don't reveal whether email exists
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(MessageResponse{
			Message: "If email exists in our system, a password reset link will be sent",
		})
		return
	}

	// Generate reset token
	resetToken, err := auth.GenerateResetToken()
	if err != nil {
		http.Error(w, "Error generating reset token", http.StatusInternalServerError)
		return
	}

	// Set reset token with expiration (1 hour)
	expiresAt := time.Now().Add(1 * time.Hour)
	_, err = h.userModel.SetPasswordResetToken(r.Context(), user.ID, resetToken, expiresAt)
	if err != nil {
		http.Error(w, "Error processing request", http.StatusInternalServerError)
		return
	}

	// Send email (placeholder)
	err = auth.SendPasswordResetEmail(user.Email, resetToken)
	if err != nil {
		http.Error(w, "Error sending email", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MessageResponse{
		Message: "If email exists in our system, a password reset link will be sent",
	})
}

// ResetPassword handler
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Token == "" || req.NewPassword == "" {
		http.Error(w, "Token and new password are required", http.StatusBadRequest)
		return
	}

	// Get user by reset token
	user, err := h.userModel.GetByResetToken(r.Context(), req.Token)
	if err != nil || user == nil {
		http.Error(w, "Invalid or expired reset token", http.StatusUnauthorized)
		return
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "Error resetting password", http.StatusInternalServerError)
		return
	}

	// Update password and clear reset token
	updatedUser, err := h.userModel.UpdatePassword(r.Context(), user.ID, hashedPassword)
	if err != nil {
		http.Error(w, "Error updating password", http.StatusInternalServerError)
		return
	}

	// Clear the reset token
	_, err = h.userModel.ClearPasswordResetToken(r.Context(), updatedUser.ID)
	if err != nil {
		http.Error(w, "Error clearing reset token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MessageResponse{
		Message: "Password has been reset successfully",
	})
}
