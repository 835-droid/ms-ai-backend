package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"math/big"
)

// charset used for invite codes
const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateRandomCode creates a cryptographically secure random code of the given length.
func GenerateRandomCode(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be > 0")
	}
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[idx.Int64()]
	}
	return string(b), nil
}

// GenerateSecureToken returns a url-safe base64 token with at least nBytes of entropy.
func GenerateSecureToken(nBytes int) (string, error) {
	if nBytes <= 0 {
		return "", errors.New("nBytes must be > 0")
	}
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
