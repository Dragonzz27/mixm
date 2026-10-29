package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Type string `json:"type"`
	jwt.RegisteredClaims
}

// Sign 签发带有令牌类型和有效期的 JWT。
func Sign(secret, subject, kind string, ttl time.Duration) (string, error) {
	now := time.Now()
	return jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		Claims{
			Type: kind,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   subject,
				IssuedAt:  jwt.NewNumericDate(now),
				ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			},
		},
	).SignedString([]byte(secret))
}

// Parse 验证 JWT 签名、过期时间、主体和令牌类型。
func Parse(secret, raw, kind string) (*Claims, error) {
	c := &Claims{}
	t, err := jwt.ParseWithClaims(
		raw,
		c,
		func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !t.Valid || c.Subject == "" || c.Type != kind {
		return nil, errors.New("invalid token")
	}

	return c, nil
}
