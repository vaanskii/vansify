package auth

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateVerificationToken generates a random verification token
func GenerateVerificationToken(email string) string {
	// Create a random byte slice
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	return base64.URLEncoding.EncodeToString(b) + ":" + email 
}