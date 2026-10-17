package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "password"
	hash, err := HashPassword(password)
	if err != nil {
		t.Errorf("Error hashing password: %v", err)
	}
	if hash == "" {
		t.Error("Expected non-empty hash")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "password"
	hash, err := HashPassword(password)
	if err != nil {
		t.Errorf("Error hashing password: %v", err)
	}
	isMatch, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Errorf("Error checking password hash: %v", err)
	}
	if !isMatch {
		t.Error("Expected password to match hash")
	}
}
