package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/furadx/iam-go/internal/apiserver/model"
	srvv1 "github.com/furadx/iam-go/internal/apiserver/service/v1"
	"github.com/furadx/iam-go/internal/pkg/code"
)

type createFakeService struct {
	users *createFakeUsers
}

func (s createFakeService) Users() srvv1.UserSrv { return s.users }

type createFakeUsers struct {
	created *model.User
}

func (u *createFakeUsers) Create(ctx context.Context, user *model.User, opts model.CreateOptions) error {
	u.created = user
	return nil
}
func (u *createFakeUsers) Update(ctx context.Context, user *model.User, opts model.UpdateOptions) error {
	return nil
}
func (u *createFakeUsers) Delete(ctx context.Context, username string, opts model.DeleteOptions) error {
	return nil
}
func (u *createFakeUsers) Get(ctx context.Context, username string, opts model.GetOptions) (*model.User, error) {
	return nil, nil
}
func (u *createFakeUsers) List(ctx context.Context, opts model.ListOptions) (*model.UserList, error) {
	return nil, nil
}
func (u *createFakeUsers) ChangePassword(ctx context.Context, username, oldPassword, newPassword string) error {
	return nil
}
func (u *createFakeUsers) Login(ctx context.Context, username, password, ip string) (*model.User, error) {
	return nil, nil
}

func postCreate(t *testing.T, body string) (*createFakeUsers, *httptest.ResponseRecorder) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	users := &createFakeUsers{}
	ctl := &UserController{srv: createFakeService{users: users}}
	r := gin.New()
	r.POST("/users", ctl.Create)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return users, w
}

// 注册不得允许客户端通过请求体设置 is_admin / status（批量赋值防护）。
func TestCreateIgnoresClientPrivilegedFields(t *testing.T) {
	users, w := postCreate(t, `{"name":"mallory","password":"Str0ng-Passw0rd!","email":"m@example.com","isAdmin":1,"status":9}`)

	if got := responseCode(t, w.Body.String()); got != code.OK {
		t.Fatalf("expected OK, got %d body=%s", got, w.Body.String())
	}
	if users.created == nil {
		t.Fatal("expected service.Create to be called")
	}
	if users.created.IsAdmin != 0 {
		t.Fatalf("client must not set is_admin, got %d", users.created.IsAdmin)
	}
	if users.created.Status != 0 {
		t.Fatalf("client must not set status, got %d", users.created.Status)
	}
}

// 注册应正确传递白名单字段。
func TestCreateMapsWhitelistedFields(t *testing.T) {
	users, w := postCreate(t, `{"name":"alice","nickname":"Alice","password":"Str0ng-Passw0rd!","email":"alice@example.com","phone":"13800138000"}`)

	if got := responseCode(t, w.Body.String()); got != code.OK {
		t.Fatalf("expected OK, got %d body=%s", got, w.Body.String())
	}
	if users.created == nil {
		t.Fatal("expected service.Create to be called")
	}
	if users.created.Name != "alice" || users.created.Nickname != "Alice" ||
		users.created.Email != "alice@example.com" || users.created.Phone != "13800138000" ||
		users.created.Password != "Str0ng-Passw0rd!" {
		t.Fatalf("whitelisted fields not mapped correctly: %+v", users.created)
	}
}

// 响应体不得包含 password 字段（H3：脱敏）。
func TestCreateResponseHidesPassword(t *testing.T) {
	_, w := postCreate(t, `{"name":"bob","password":"Str0ng-Passw0rd!","email":"bob@example.com"}`)
	if strings.Contains(w.Body.String(), "Str0ng-Passw0rd!") || strings.Contains(strings.ToLower(w.Body.String()), "\"password\"") {
		t.Fatalf("response must not leak password, body=%s", w.Body.String())
	}
}
