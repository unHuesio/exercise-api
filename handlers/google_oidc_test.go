package handlers

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
)

func newTestGoogleOIDCProvider(t *testing.T) (*httptest.Server, *rsa.PrivateKey, string) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"issuer":                 serverURL,
				"jwks_uri":               serverURL + "/keys",
				"authorization_endpoint": serverURL + "/authorize",
				"token_endpoint":         serverURL + "/token",
			})
		case "/keys":
			jwk := jose.JSONWebKey{
				Key:       &privateKey.PublicKey,
				KeyID:     "test-kid",
				Algorithm: "RS256",
				Use:       "sig",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"keys": []jose.JSONWebKey{jwk.Public()}})
		default:
			http.NotFound(w, r)
		}
	}))
	serverURL = server.URL

	return server, privateKey, server.URL
}

func TestValidateGoogleTokenClaimsRejectsInvalidInputs(t *testing.T) {
	t.Run("missing subject", func(t *testing.T) {
		err := ValidateGoogleTokenClaims(GoogleTokenClaims{Email: "user@example.com", EmailVerified: true, Audience: "client-123", Issuer: "https://accounts.google.com", Expiry: time.Now().Add(time.Hour).Unix()}, "client-123", "https://accounts.google.com")
		if err == nil {
			t.Fatal("expected error for missing subject")
		}
	})

	t.Run("email not verified", func(t *testing.T) {
		err := ValidateGoogleTokenClaims(GoogleTokenClaims{Subject: "123", Email: "user@example.com", EmailVerified: false, Audience: "client-123", Issuer: "https://accounts.google.com", Expiry: time.Now().Add(time.Hour).Unix()}, "client-123", "https://accounts.google.com")
		if err == nil {
			t.Fatal("expected error for unverified email")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		err := ValidateGoogleTokenClaims(GoogleTokenClaims{Subject: "123", Email: "user@example.com", EmailVerified: true, Audience: "client-123", Issuer: "https://accounts.google.com", Expiry: time.Now().Add(-time.Minute).Unix()}, "client-123", "https://accounts.google.com")
		if err == nil {
			t.Fatal("expected error for expired token")
		}
	})
}

func TestVerifyGoogleIDTokenAcceptsValidGoogleToken(t *testing.T) {
	server, privateKey, issuerURL := newTestGoogleOIDCProvider(t)
	defer server.Close()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss":            issuerURL,
		"sub":            "google-user-123",
		"aud":            "client-123",
		"email":          "user@example.com",
		"email_verified": true,
		"exp":            time.Now().Add(time.Hour).Unix(),
	})
	token.Header["kid"] = "test-kid"
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	claims, err := VerifyGoogleIDToken(context.Background(), tokenString, "client-123", issuerURL)
	if err != nil {
		t.Fatalf("expected valid token to verify, got error: %v", err)
	}
	if claims.Subject != "google-user-123" {
		t.Fatalf("expected subject google-user-123, got %s", claims.Subject)
	}
	if claims.Email != "user@example.com" {
		t.Fatalf("expected email user@example.com, got %s", claims.Email)
	}
}
