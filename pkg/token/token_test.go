package token

import (
	"errors"
	"testing"
	"time"
)

func newTestManager() *Manager {
	return NewManager("test-secret", 15*time.Minute, 7*24*time.Hour)
}

func TestSignAccessAndParse(t *testing.T) {
	m := newTestManager()
	tok, claims, err := m.SignAccess(42, "alice")
	if err != nil {
		t.Fatalf("SignAccess failed: %v", err)
	}
	if claims.Type != TypeAccess || claims.UserID != 42 || claims.Username != "alice" || claims.ID == "" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	got, err := m.Parse(tok)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if got.UserID != 42 || got.Username != "alice" || got.Type != TypeAccess {
		t.Fatalf("parsed claims mismatch: %+v", got)
	}
}

func TestParseTypedRejectsWrongType(t *testing.T) {
	m := newTestManager()
	refresh, _, err := m.SignRefresh(1, "bob")
	if err != nil {
		t.Fatalf("SignRefresh failed: %v", err)
	}
	if _, err := m.ParseTyped(refresh, TypeAccess); !errors.Is(err, ErrWrongTokenType) {
		t.Fatalf("expected ErrWrongTokenType, got %v", err)
	}
	if _, err := m.ParseTyped(refresh, TypeRefresh); err != nil {
		t.Fatalf("ParseTyped(refresh) should pass, got %v", err)
	}
}

func TestParseExpired(t *testing.T) {
	m := NewManager("test-secret", -time.Hour, -time.Hour)
	tok, _, err := m.SignAccess(1, "x")
	if err != nil {
		t.Fatalf("SignAccess failed: %v", err)
	}
	if _, err := m.Parse(tok); !errors.Is(err, ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}

func TestParseWrongSecret(t *testing.T) {
	signer := NewManager("secret-a", time.Hour, time.Hour)
	verifier := NewManager("secret-b", time.Hour, time.Hour)
	tok, _, _ := signer.SignAccess(1, "x")
	if _, err := verifier.Parse(tok); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestParseMalformed(t *testing.T) {
	m := newTestManager()
	if _, err := m.Parse("not-a-jwt"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAccessExpireSeconds(t *testing.T) {
	m := NewManager("s", 15*time.Minute, time.Hour)
	if m.AccessExpireSeconds() != 900 {
		t.Fatalf("expected 900, got %d", m.AccessExpireSeconds())
	}
}
