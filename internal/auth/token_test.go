package auth

import (
	"testing"
	"time"
)

func TestGenerateAndValidateToken(t *testing.T) {
	secret := []byte("test-cluster-secret")

	token, err := GenerateJoinToken(secret, 24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	valid, err := ValidateJoinToken(token, secret)
	if err != nil {
		t.Fatalf("validation error: %v", err)
	}
	if !valid {
		t.Error("expected token to be valid")
	}
}

func TestExpiredToken(t *testing.T) {
	secret := []byte("test-secret")

	token, err := GenerateJoinToken(secret, -1*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	valid, err := ValidateJoinToken(token, secret)
	if err == nil || valid {
		t.Error("expected expired token to be invalid")
	}
}

func TestInvalidToken(t *testing.T) {
	secret := []byte("test-secret")

	valid, err := ValidateJoinToken("garbage-token", secret)
	if err == nil || valid {
		t.Error("expected invalid token to fail")
	}
}
