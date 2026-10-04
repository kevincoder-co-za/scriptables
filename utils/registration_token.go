package utils

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"os"
	"strings"
)

const defaultRegistrationTokenFile = "/etc/secrets/scriptables/registration_token"

func registrationTokenFile() string {
	if path := os.Getenv("REGISTRATION_TOKEN_FILE"); path != "" {
		return path
	}

	return defaultRegistrationTokenFile
}

func readRegistrationTokenHash() string {
	hash, err := os.ReadFile(registrationTokenFile())
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(hash))
}

func HasUnusedRegistrationToken() bool {
	return readRegistrationTokenHash() != ""
}

func IsValidRegistrationToken(token string) bool {
	storedHash := readRegistrationTokenHash()
	if storedHash == "" {
		return false
	}

	tokenHash := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return subtle.ConstantTimeCompare([]byte(hex.EncodeToString(tokenHash[:])), []byte(storedHash)) == 1
}

func BurnRegistrationToken() error {
	return os.Remove(registrationTokenFile())
}
