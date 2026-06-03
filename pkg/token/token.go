package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 哨兵错误，供上层（中间件）映射为业务错误码，避免 pkg 反向依赖 internal。
var (
	// ErrInvalidToken 表示 Token 签名无效、格式错误或校验失败。
	ErrInvalidToken = errors.New("invalid token")
	// ErrTokenExpired 表示 Token 已过期。
	ErrTokenExpired = errors.New("token expired")
)

// Claims 自定义 JWT 声明，携带用户 ID。
type Claims struct {
	UserID int64 `json:"uid"`
	jwt.RegisteredClaims
}

// Manager 负责 JWT 的签发与校验。
type Manager struct {
	secret []byte
	expire time.Duration
	issuer string
}

// NewManager 创建 Token 管理器。expire 为 Token 有效期。
func NewManager(secret string, expire time.Duration) *Manager {
	return &Manager{
		secret: []byte(secret),
		expire: expire,
		issuer: "iam-go",
	}
}

// Sign 为指定用户签发一个 HS256 JWT。
func (m *Manager) Sign(userID int64) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expire)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Parse 校验并解析 JWT，返回用户 ID。
func (m *Manager) Parse(tokenStr string) (int64, error) {
	var claims Claims
	tkn, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		// 只接受 HMAC 签名，防止算法混淆攻击（alg=none / RS256 伪造）。
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return 0, ErrTokenExpired
		}
		return 0, ErrInvalidToken
	}
	if !tkn.Valid {
		return 0, ErrInvalidToken
	}
	return claims.UserID, nil
}
