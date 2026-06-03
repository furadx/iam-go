package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// 令牌类型。
const (
	TypeAccess  = "access"
	TypeRefresh = "refresh"
)

// 哨兵错误，供上层映射为业务错误码，避免 pkg 反向依赖 internal。
var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrTokenExpired   = errors.New("token expired")
	ErrWrongTokenType = errors.New("wrong token type")
)

// Claims 自定义 JWT 声明。
type Claims struct {
	UserID   int64  `json:"uid"`
	Username string `json:"username"`
	Type     string `json:"typ"`
	jwt.RegisteredClaims
}

// Manager 负责 JWT 的签发与校验。
type Manager struct {
	secret        []byte
	accessExpire  time.Duration
	refreshExpire time.Duration
	issuer        string
}

// NewManager 创建 Token 管理器。
func NewManager(secret string, accessExpire, refreshExpire time.Duration) *Manager {
	return &Manager{
		secret:        []byte(secret),
		accessExpire:  accessExpire,
		refreshExpire: refreshExpire,
		issuer:        "iam-go",
	}
}

// AccessExpireSeconds 返回 access token 有效期（秒），用于登录响应 expires_in。
func (m *Manager) AccessExpireSeconds() int {
	return int(m.accessExpire.Seconds())
}

func (m *Manager) sign(uid int64, username, typ string, expire time.Duration) (string, *Claims, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   uid,
		Username: username,
		Type:     typ,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expire)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", nil, err
	}
	return signed, claims, nil
}

// SignAccess 签发 access token。
func (m *Manager) SignAccess(uid int64, username string) (string, *Claims, error) {
	return m.sign(uid, username, TypeAccess, m.accessExpire)
}

// SignRefresh 签发 refresh token。
func (m *Manager) SignRefresh(uid int64, username string) (string, *Claims, error) {
	return m.sign(uid, username, TypeRefresh, m.refreshExpire)
}

// Parse 校验签名与过期，返回 claims。
func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	var claims Claims
	tkn, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}
	if !tkn.Valid {
		return nil, ErrInvalidToken
	}
	return &claims, nil
}

// ParseTyped 在 Parse 基础上额外校验令牌类型。
func (m *Manager) ParseTyped(tokenStr, want string) (*Claims, error) {
	claims, err := m.Parse(tokenStr)
	if err != nil {
		return nil, err
	}
	if claims.Type != want {
		return nil, ErrWrongTokenType
	}
	return claims, nil
}
