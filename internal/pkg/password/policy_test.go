package password

import (
	"errors"
	"testing"
)

func TestPolicyValidateAcceptsStrongPassword(t *testing.T) {
	policy := DefaultPolicy()
	if err := policy.Validate("alice", "Str0ng!Passw0rd"); err != nil {
		t.Fatalf("expected strong password to pass, got %v", err)
	}
}

func TestPolicyValidateRejectsShortPassword(t *testing.T) {
	policy := DefaultPolicy()
	if err := policy.Validate("alice", "A1!short"); !errors.Is(err, ErrTooShort) {
		t.Fatalf("expected ErrTooShort, got %v", err)
	}
}

func TestPolicyValidateRejectsUsername(t *testing.T) {
	policy := DefaultPolicy()
	if err := policy.Validate("alice", "Alice123456!"); !errors.Is(err, ErrContainsUsername) {
		t.Fatalf("expected ErrContainsUsername, got %v", err)
	}
}

func TestPolicyValidateRejectsTooFewClasses(t *testing.T) {
	policy := DefaultPolicy()
	if err := policy.Validate("alice", "lowercaseonly"); !errors.Is(err, ErrTooFewClasses) {
		t.Fatalf("expected ErrTooFewClasses, got %v", err)
	}
}

func TestPolicyValidateRejectsCommonPassword(t *testing.T) {
	policy := DefaultPolicy()
	if err := policy.Validate("alice", "password1234"); !errors.Is(err, ErrCommonPassword) {
		t.Fatalf("expected ErrCommonPassword, got %v", err)
	}
}
