package token

import (
	"errors"
	"testing"
	"time"
)

func TestSignAndParse(t *testing.T) {
	m := NewManager("test-secret", time.Hour)

	tokenStr, err := m.Sign(42)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	userID, err := m.Parse(tokenStr)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if userID != 42 {
		t.Fatalf("expected userID 42, got %d", userID)
	}
}

func TestParseExpired(t *testing.T) {
	m := NewManager("test-secret", -time.Hour) // 已过期
	tokenStr, err := m.Sign(1)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	if _, err := m.Parse(tokenStr); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}

func TestParseWrongSecret(t *testing.T) {
	signer := NewManager("secret-a", time.Hour)
	verifier := NewManager("secret-b", time.Hour)

	tokenStr, err := signer.Sign(1)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	if _, err := verifier.Parse(tokenStr); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestParseMalformed(t *testing.T) {
	m := NewManager("test-secret", time.Hour)

	if _, err := m.Parse("not-a-jwt"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}
