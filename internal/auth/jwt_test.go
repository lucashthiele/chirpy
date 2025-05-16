package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	secret := "testsecret"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}
	if token == "" {
		t.Error("MakeJWT returned empty token")
	}
}

func TestValidateJwt(t *testing.T) {
	userID := uuid.New()
	secret := "testsecret"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	parsedID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT returned error: %v", err)
	}
	if parsedID != userID {
		t.Errorf("ValidateJWT returned wrong userID: got %v, want %v", parsedID, userID)
	}

	// Test with wrong secret
	_, err = ValidateJWT(token, "wrongsecret")
	if err == nil {
		t.Error("ValidateJWT should fail with wrong secret")
	}
}
