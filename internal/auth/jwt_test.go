package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	secret := "testsecret"
	expiresIn := time.Hour

	cases := []struct {
		name      string
		userID    uuid.UUID
		secret    string
		expiresIn time.Duration
		wantErr   bool
	}{
		{
			name:      "valid input",
			userID:    userID,
			secret:    secret,
			expiresIn: expiresIn,
			wantErr:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := MakeJWT(tc.userID, tc.secret, tc.expiresIn)
			if tc.wantErr {
				if err == nil {
					t.Error("MakeJWT should fail but did not")
				}
			} else {
				if err != nil {
					t.Fatalf("MakeJWT returned error: %v", err)
				}
				if token == "" {
					t.Error("MakeJWT returned empty token")
				}
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "testsecret"
	expiresIn := time.Hour
	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	cases := []struct {
		name       string
		token      string
		secret     string
		wantErr    bool
		wantUserID uuid.UUID
	}{
		{
			name:       "valid token and secret",
			token:      token,
			secret:     secret,
			wantErr:    false,
			wantUserID: userID,
		},
		{
			name:       "invalid secret",
			token:      token,
			secret:     "wrongsecret",
			wantErr:    true,
			wantUserID: uuid.Nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parsedID, err := ValidateJWT(tc.token, tc.secret)
			if tc.wantErr {
				if err == nil {
					t.Error("ValidateJWT should fail but did not")
				}
			} else {
				if err != nil {
					t.Fatalf("ValidateJWT returned error: %v", err)
				}
				if parsedID != tc.wantUserID {
					t.Errorf("ValidateJWT returned wrong userID: got %v, want %v", parsedID, tc.wantUserID)
				}
			}
		})
	}
}

func TestGetBearerToken(t *testing.T) {
	cases := []struct {
		name    string
		headers map[string]string
		want    string
		wantErr bool
	}{
		{
			name:    "valid bearer token",
			headers: map[string]string{"Authorization": "Bearer sometoken"},
			want:    "sometoken",
			wantErr: false,
		},
		{
			name:    "missing authorization header",
			headers: map[string]string{},
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty token",
			headers: map[string]string{"Authorization": "Bearer "},
			want:    "",
			wantErr: true,
		},
		{
			name:    "no bearer prefix",
			headers: map[string]string{"Authorization": "sometoken"},
			want:    "sometoken",
			wantErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := make(http.Header)
			for k, v := range tc.headers {
				h.Set(k, v)
			}
			token, err := GetBearerToken(&h)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if token != tc.want {
					t.Errorf("got token %q, want %q", token, tc.want)
				}
			}
		})
	}
}
