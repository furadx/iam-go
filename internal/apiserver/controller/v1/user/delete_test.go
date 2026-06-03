package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/apiserver/model"
	srvv1 "github.com/furadx/iam-go/internal/apiserver/service/v1"
	"github.com/furadx/iam-go/internal/pkg/code"
)

type deleteFakeService struct {
	users *deleteFakeUsers
}

func (s deleteFakeService) Users() srvv1.UserSrv {
	return s.users
}

type deleteFakeUsers struct {
	deleted string
}

func (u *deleteFakeUsers) Create(ctx context.Context, user *model.User, opts model.CreateOptions) error {
	return nil
}
func (u *deleteFakeUsers) Update(ctx context.Context, user *model.User, opts model.UpdateOptions) error {
	return nil
}
func (u *deleteFakeUsers) Delete(ctx context.Context, username string, opts model.DeleteOptions) error {
	u.deleted = username
	return nil
}
func (u *deleteFakeUsers) Get(ctx context.Context, username string, opts model.GetOptions) (*model.User, error) {
	return nil, nil
}
func (u *deleteFakeUsers) List(ctx context.Context, opts model.ListOptions) (*model.UserList, error) {
	return nil, nil
}
func (u *deleteFakeUsers) ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	return nil
}
func (u *deleteFakeUsers) Login(ctx context.Context, username, password, ip string) (*model.User, error) {
	return nil, nil
}

func TestDeleteCallsServiceWithRouteName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	users := &deleteFakeUsers{}
	ctl := &UserController{srv: deleteFakeService{users: users}}
	r := gin.New()
	r.DELETE("/users/:name", ctl.Delete)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/users/alice", nil)
	r.ServeHTTP(w, req)

	if got := responseCode(t, w.Body.String()); got != code.OK {
		t.Fatalf("expected OK, got %d body=%s", got, w.Body.String())
	}
	if users.deleted != "alice" {
		t.Fatalf("expected service delete for alice, got %q", users.deleted)
	}
}
