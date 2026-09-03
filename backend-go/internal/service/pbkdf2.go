package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/crypto/pbkdf2"
)

// hashPasswordPBKDF2 hashes a password using PBKDF2WithHmacSHA256.
// Format: Base64(salt):Base64(hash)
func hashPasswordPBKDF2(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	iterations := 65536
	keyLen := 32
	dk := pbkdf2.Key([]byte(password), salt, iterations, keyLen, sha256.New)
	return base64.StdEncoding.EncodeToString(salt) + ":" + base64.StdEncoding.EncodeToString(dk), nil
}

// checkPasswordPBKDF2 verifies a raw password against the stored PBKDF2 format.
func checkPasswordPBKDF2(password, stored string) bool {
	if stored == "" {
		return false
	}
	// split on first ':'
	idx := -1
	for i := 0; i < len(stored); i++ {
		if stored[i] == ':' {
			idx = i
			break
		}
	}
	if idx < 0 {
		return false
	}
	saltPart := stored[:idx]
	hashPart := stored[idx+1:]
	saltB, err := base64.StdEncoding.DecodeString(saltPart)
	if err != nil {
		return false
	}
	iterations := 65536
	keyLen := 32
	dk := pbkdf2.Key([]byte(password), saltB, iterations, keyLen, sha256.New)
	computed := base64.StdEncoding.EncodeToString(dk)
	return computed == hashPart
}
