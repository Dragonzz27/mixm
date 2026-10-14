package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenBoundaries(t *testing.T) {
	for _, kind := range []string{"access", "refresh"} {
		raw, err := Sign("test-secret", "learner", kind, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		claims, err := Parse("test-secret", raw, kind)
		if err != nil || claims.Subject != "learner" {
			t.Fatalf("valid token rejected: %v", err)
		}
		other := "access"
		if kind == "access" {
			other = "refresh"
		}
		if _, err = Parse("test-secret", raw, other); err == nil {
			t.Fatal("wrong token type accepted")
		}
		if _, err = Parse("wrong-secret", raw, kind); err == nil {
			t.Fatal("wrong signature accepted")
		}
	}
	expired, _ := Sign("test-secret", "learner", "access", -time.Minute)
	if _, err := Parse("test-secret", expired, "access"); err == nil {
		t.Fatal("expired token accepted")
	}
	unsigned, err := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{Type: "access", RegisteredClaims: jwt.RegisteredClaims{Subject: "learner", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Parse("test-secret", unsigned, "access"); err == nil {
		t.Fatal("unsigned token accepted")
	}
}
func TestPasswordHash(t *testing.T) {
	hash, err := Hash("课程-password")
	if err != nil {
		t.Fatal(err)
	}
	if !Verify(hash, "课程-password") || Verify(hash, "wrong") {
		t.Fatal("password verification failed")
	}
}

func TestTokenRequiresType(t *testing.T) {
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: "learner", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}).SignedString([]byte("secret"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Parse("secret", raw, "access"); err == nil {
		t.Fatal("access token without type accepted")
	}
	if _, err = Parse("secret", raw, "refresh"); err == nil {
		t.Fatal("refresh token without type accepted")
	}
}
