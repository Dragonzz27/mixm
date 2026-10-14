package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenBoundaries(t *testing.T) {
	for _, kind := range []string{"access", "refresh"} {
		raw, e := Sign("test-secret", "learner", kind, time.Minute)
		if e != nil {
			t.Fatal(e)
		}
		claims, e := Parse("test-secret", raw, kind)
		if e != nil || claims.Subject != "learner" {
			t.Fatalf("valid token rejected: %v", e)
		}
		other := "access"
		if kind == "access" {
			other = "refresh"
		}
		if _, e = Parse("test-secret", raw, other); e == nil {
			t.Fatal("wrong token type accepted")
		}
		if _, e = Parse("wrong-secret", raw, kind); e == nil {
			t.Fatal("wrong signature accepted")
		}
	}
	expired, _ := Sign("test-secret", "learner", "access", -time.Minute)
	if _, e := Parse("test-secret", expired, "access"); e == nil {
		t.Fatal("expired token accepted")
	}
	unsigned, e := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{Type: "access", RegisteredClaims: jwt.RegisteredClaims{Subject: "learner", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Parse("test-secret", unsigned, "access"); e == nil {
		t.Fatal("unsigned token accepted")
	}
}
func TestBcryptCompatibility(t *testing.T) {
	hash, e := Hash("课程-password")
	if e != nil {
		t.Fatal(e)
	}
	if !Verify(hash, "课程-password") || Verify(hash, "wrong") {
		t.Fatal("password verification failed")
	}
	pythonHash := "$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6Ttx2zZYPcLdQxrpVElo5VS1s.CGa"
	if Verify(pythonHash, "wrong") {
		t.Fatal("invalid legacy password accepted")
	}
}

func TestLegacyAccessToken(t *testing.T) {
	raw, e := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: "legacy", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}).SignedString([]byte("secret"))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Parse("secret", raw, "access"); e != nil {
		t.Fatal("legacy access token rejected", e)
	}
	if _, e = Parse("secret", raw, "refresh"); e == nil {
		t.Fatal("legacy access token accepted as refresh")
	}
}
