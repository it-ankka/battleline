package gameutils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"strings"
)

func GenerateID(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	id := base64.URLEncoding.EncodeToString(bytes)
	return strings.TrimRight(id[:length], "="), nil
}

func GenerateKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
