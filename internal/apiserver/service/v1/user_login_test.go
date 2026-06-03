package v1

import (
	"context"
	"errors"
	"testing"

	"github.com/furadx/iam-go/internal/apiserver/model"
	"github.com/furadx/iam-go/internal/apiserver/store"
	"github.com/furadx/iam-go/internal/pkg/code"
	"github.com/furadx/iam-go/internal/pkg/loginlock"
	"github.com/furadx/iam-go/internal/pkg/password"
)

type fakeFactory struct {
	users *fakeUserStore
}

func (f *fakeFactory) Users() store.UserStore { return f.users }
func (f *fakeFactory) Close() error           { return nil }

type fakeUserStore struct {
	getCalls int
	getErr   error
	user     *model.User
}

func (s *fakeUserStore) Create(ctx context.Context, user *model.User, opts model.CreateOptions) error {
	return nil
}

func (s *fakeUserStore) Update(ctx context.Context, user *model.User, opts model.UpdateOptions) error {
	return nil
}

func (s *fakeUserStore) Delete(ctx context.Context, username string, opts model.DeleteOptions) error {
	return nil
}

func (s *fakeUserStore) Get(ctx context.Context, username string, opts model.GetOptions) (*model.User, error) {
	s.getCalls++
	if s.getErr != nil {
		return nil, s.getErr
	}
	return s.user, nil
}

func (s *fakeUserStore) List(ctx context.Context, opts model.ListOptions) (*model.UserList, error) {
	return nil, nil
}

type fakeLoginGuard struct {
	checkErr       error
	recordFailures int
	resets         int
}

func (g *fakeLoginGuard) Check(ctx context.Context, username, ip string) error {
	return g.checkErr
}

func (g *fakeLoginGuard) RecordFailure(ctx context.Context, username, ip string) error {
	g.recordFailures++
	return nil
}

func (g *fakeLoginGuard) Reset(ctx context.Context, username, ip string) error {
	g.resets++
	return nil
}

func TestLoginReturnsLockedBeforeStoreLookup(t *testing.T) {
	users := &fakeUserStore{}
	guard := &fakeLoginGuard{checkErr: loginlock.ErrLocked}
	svc := NewService(&fakeFactory{users: users}, nil, nil, guard, password.DefaultPolicy())

	_, err := svc.Users().Login(context.Background(), "alice", "wrong", "127.0.0.1")
	if code.Code(err) != code.ErrLoginLocked {
		t.Fatalf("expected ErrLoginLocked, got %v", err)
	}
	if users.getCalls != 0 {
		t.Fatalf("expected login lock to short-circuit store lookup, got %d calls", users.getCalls)
	}
}

func TestLoginHidesUserNotFoundAndRecordsFailure(t *testing.T) {
	users := &fakeUserStore{getErr: code.New(code.ErrUserNotFound)}
	guard := &fakeLoginGuard{}
	svc := NewService(&fakeFactory{users: users}, nil, nil, guard, password.DefaultPolicy())

	_, err := svc.Users().Login(context.Background(), "missing", "wrong", "127.0.0.1")
	if code.Code(err) != code.ErrPasswordIncorrect {
		t.Fatalf("expected ErrPasswordIncorrect, got %v", err)
	}
	if guard.recordFailures != 1 {
		t.Fatalf("expected one recorded failure, got %d", guard.recordFailures)
	}
}

func TestLoginHidesDisabledUserAndRecordsFailure(t *testing.T) {
	users := &fakeUserStore{user: &model.User{Name: "alice", Status: 0}}
	guard := &fakeLoginGuard{}
	svc := NewService(&fakeFactory{users: users}, nil, nil, guard, password.DefaultPolicy())

	_, err := svc.Users().Login(context.Background(), "alice", "wrong", "127.0.0.1")
	if code.Code(err) != code.ErrPasswordIncorrect {
		t.Fatalf("expected ErrPasswordIncorrect, got %v", err)
	}
	if guard.recordFailures != 1 {
		t.Fatalf("expected one recorded failure, got %d", guard.recordFailures)
	}
	if guard.resets != 0 {
		t.Fatalf("expected no failure reset, got %d", guard.resets)
	}
}

func TestLoginIgnoresLoginGuardStoreErrors(t *testing.T) {
	users := &fakeUserStore{getErr: errors.New("database down")}
	guard := &fakeLoginGuard{checkErr: context.DeadlineExceeded}
	svc := NewService(&fakeFactory{users: users}, nil, nil, guard, password.DefaultPolicy())

	_, err := svc.Users().Login(context.Background(), "alice", "wrong", "127.0.0.1")
	if err == nil || code.Code(err) != code.ErrInternal {
		t.Fatalf("expected database error to continue surfacing as internal error, got %v", err)
	}
	if users.getCalls != 1 {
		t.Fatalf("expected store lookup despite login guard error, got %d calls", users.getCalls)
	}
}
