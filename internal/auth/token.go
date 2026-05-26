package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

type tokenPayload struct {
	Nonce     string    `json:"nonce"`
	ExpiresAt time.Time `json:"expires_at"`
}

func GenerateJoinToken(secret []byte, ttl time.Duration) (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}

	payload := tokenPayload{
		Nonce:     base64.RawURLEncoding.EncodeToString(nonce),
		ExpiresAt: time.Now().Add(ttl),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshaling payload: %w", err)
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write(payloadBytes)
	sig := mac.Sum(nil)

	token := base64.RawURLEncoding.EncodeToString(payloadBytes) + "." + base64.RawURLEncoding.EncodeToString(sig)
	return "kudo-v1:" + token, nil
}

func ValidateJoinToken(token string, secret []byte) (bool, error) {
	if len(token) < 8 || token[:8] != "kudo-v1:" {
		return false, fmt.Errorf("invalid token format")
	}

	parts := splitToken(token[8:])
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid token structure")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false, fmt.Errorf("decoding payload: %w", err)
	}

	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return false, fmt.Errorf("decoding signature: %w", err)
	}

	mac := hmac.New(sha256.New, secret)
	mac.Write(payloadBytes)
	expectedSig := mac.Sum(nil)
	if !hmac.Equal(sig, expectedSig) {
		return false, fmt.Errorf("invalid signature")
	}

	var payload tokenPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return false, fmt.Errorf("parsing payload: %w", err)
	}

	if time.Now().After(payload.ExpiresAt) {
		return false, fmt.Errorf("token expired")
	}

	return true, nil
}

func splitToken(s string) []string {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return []string{s[:i], s[i+1:]}
		}
	}
	return []string{s}
}
