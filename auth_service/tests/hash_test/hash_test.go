package hash_test

import (
	"testing"
	"auth_service/pkg/hash"
)

func TestHashPassword_Success(t *testing.T) {
	password := "secret123"

	hashed, err := hash.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	if hashed == password {
		t.Error("Hash must not equal original password")
	}
	if len(hashed) == 0 {
		t.Error("Hash must not be empty")
	}
}

func TestCheckPasswordHash_Valid(t *testing.T) {
	password := "secret123"
	hashed, _ := hash.HashPassword(password)

	if !hash.CheckPasswordHash(password, hashed) {
		t.Error("Valid password should match hash")
	}
}

func TestCheckPasswordHash_Invalid(t *testing.T) {
	password := "secret123"
	hashed, _ := hash.HashPassword(password)

	if hash.CheckPasswordHash("wrongpassword", hashed) {
		t.Error("Wrong password should not match hash")
	}
}