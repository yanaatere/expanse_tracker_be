package auth

import (
	"crypto/rand"
	"encoding/hex"
	"log"
)

func GenerateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func SendPasswordResetEmail(email, resetToken string) error {
	// TODO: integrate with an email provider (SendGrid, Mailgun, AWS SES).
	// The reset token is intentionally NOT logged here to avoid leaking it to log aggregators.
	log.Printf("WARN: password reset email stub called for %s — wire up a real email provider", email)
	_ = resetToken
	return nil
}
