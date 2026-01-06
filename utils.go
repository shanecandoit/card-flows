package main

import (
	"crypto/rand"
	"encoding/hex"
)

// NewID generates a simple random unique identifier
func NewID() string {
	b := make([]byte, 8) // 16 characters hex
	if _, err := rand.Read(b); err != nil {
		return "error-id"
	}
	return hex.EncodeToString(b)
}
