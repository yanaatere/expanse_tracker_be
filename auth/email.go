package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func SendPasswordResetEmail(email, resetToken string) error {
	// This is a placeholder implementation
	// In production, integrate with an email service like SendGrid, Mailgun, or AWS SES
	resetLink := fmt.Sprintf("https://your-frontend.com/reset-password?token=%s", resetToken)
	fmt.Printf("Password reset link for %s: %s\n", email, resetLink)
	return nil
}
