package auth

import (
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
)

func TestCreateHash(t *testing.T) {
	hash1, err := argon2id.CreateHash("pa$$word", argon2id.DefaultParams)
	if err != nil {
		t.Fatal(err)
	}
	hashRegex := regexp.MustCompile(`^\$argon2id\$v=19\$m=\d+,t=\d+,p=\d+\$[A-Za-z0-9+/]+\$[A-Za-z0-9+/]+$`)
	if !hashRegex.MatchString(hash1) {
		t.Errorf("Hash %q is not in the correct format", hash1)
	}
	hash2, err := argon2id.CreateHash("pa$$word", argon2id.DefaultParams)
	if err != nil {
		t.Fatal(err)
	}
	if hash1 == hash2 {
		t.Error("Hashes must be unique due to random salts")
	}
}

func TestComparePasswordAndHash(t *testing.T) {
	password := "correct_password"
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		t.Fatal(err)
	}
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		t.Errorf("Verification error: %v", err)
	}
	if !match {
		t.Error("Expected password to match its own hash")
	}
	match, err = argon2id.ComparePasswordAndHash("wrong_password", hash)
	if err != nil {
		t.Errorf("Verification errpr: %v", err)
	}
	if match {
		t.Error("Expected incorrect password to fail verification")
	}
}

func TestValidateJWT(t *testing.T) {
	secret := "test-secret"
	userID := uuid.New()

	validToken, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to create valid token: %v", err)
	}
	expiredToken, err := MakeJWT(userID, secret, -time.Hour)
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	tests := []struct {
		name        string
		tokenString string
		secret      string
		wantID      uuid.UUID
		wantErr     bool
	}{
		{
			name:        "valid token",
			tokenString: validToken,
			secret:      secret,
			wantID:      userID,
			wantErr:     false,
		},
		{
			name:        "expired token",
			tokenString: expiredToken,
			secret:      secret,
			wantID:      uuid.Nil,
			wantErr:     true,
		},
		{
			name:        "wrong secret",
			tokenString: validToken,
			secret:      "wrong-secret",
			wantID:      uuid.Nil,
			wantErr:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotID, err := ValidateJWT(test.tokenString, test.secret)
			if (err != nil) != test.wantErr {
				t.Errorf("wantErr: %v, got err: %v", test.wantErr, err)
			}
			if gotID != test.wantID {
				t.Errorf("wantID: %v, got: %v", test.wantID, gotID)
			}
		})
	}

}

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		headers   http.Header
		wantToken string
		wantErr   bool
	}{
		{
			name:      "valid bearer token",
			headers:   http.Header{"Authorization": []string{"Bearer my-token"}},
			wantToken: "my-token",
			wantErr:   false,
		},
		{
			name:      "missing authorization header",
			headers:   http.Header{},
			wantToken: "",
			wantErr:   true,
		},
		{
			name:      "wrong prefix",
			headers:   http.Header{"Authorization": []string{"Token my-token"}},
			wantToken: "",
			wantErr:   true,
		},
		{
			name:      "empty bearer value",
			headers:   http.Header{"Authorization": []string{"Bearer "}},
			wantToken: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := GetBearerToken(tt.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr %v, got err %v", tt.wantErr, err)
			}
			if gotToken != tt.wantToken {
				t.Errorf("wantToken %v, got %v", tt.wantToken, gotToken)
			}
		})
	}
}
