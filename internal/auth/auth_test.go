package auth

import (
	"regexp"
	"testing"

	"github.com/alexedwards/argon2id"
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
