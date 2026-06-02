package jwt

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/furadx/iam-go/pkg/code"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// Manager JWT 管理器，负责生成和解析 JWT token
type Manager struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

// Claims 定义了 JWT 的有效载荷结构
type Claims struct {
	UserID int64 `json:"uid"`
	jwtv5.RegisteredClaims
}

// NewManager 创建一个新的 JWT 管理器实例
// secret: 用于签名的密钥
// issuer: token 签发者标识
// ttlSeconds: token 有效期（秒）
func NewManager(secret string, issuer string, ttlSeconds int64) *Manager {
	return &Manager{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    time.Duration(ttlSeconds) * time.Second,
	}
}

// Sign 为指定用户 ID 生成 JWT token
// 返回: token 字符串, 过期时间（秒）, 错误
func (m *Manager) Sign(userID int64) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(m.ttl)

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwtv5.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(expiresAt),
		},
	}

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", 0, err
	}

	return signed, int64(m.ttl.Seconds()), nil
}

// Parse 解析并验证 JWT token
// 返回: Claims 结构, 错误
func (m *Manager) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwtv5.ParseWithClaims(tokenString, claims, func(token *jwtv5.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtv5.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("非预期的签名算法: %v", token.Header["alg"])
		}
		return m.secret, nil
	})
	if err != nil {
		if errors.Is(err, jwtv5.ErrTokenExpired) {
			return nil, code.Wrap(code.ErrExpiredToken, err)
		}
		return nil, code.Wrap(code.ErrInvalidToken, err)
	}

	if token == nil || !token.Valid {
		return nil, code.New(code.ErrInvalidToken)
	}

	if claims.Issuer != m.issuer {
		return nil, code.New(code.ErrInvalidToken)
	}

	if claims.UserID <= 0 {
		// 兼容只使用 sub 的情况
		uid, parseErr := strconv.ParseInt(claims.Subject, 10, 64)
		if parseErr != nil || uid <= 0 {
			return nil, code.New(code.ErrInvalidToken)
		}
		claims.UserID = uid
	}

	return claims, nil
}
