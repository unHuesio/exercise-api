package handlers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
)

type GoogleTokenClaims struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Issuer        string `json:"iss"`
	Audience      string `json:"aud"`
	Expiry        int64  `json:"exp"`
}

func NewGoogleTokenVerifier(ctx context.Context, clientID, issuer string) (*oidc.IDTokenVerifier, error) {
	if clientID == "" {
		return nil, errors.New("google client id is not configured")
	}
	if issuer == "" {
		return nil, errors.New("google issuer is not configured")
	}
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}
	return provider.Verifier(&oidc.Config{ClientID: clientID}), nil
}

func VerifyGoogleIDTokenWithVerifier(ctx context.Context, verifier *oidc.IDTokenVerifier, rawToken, clientID, issuer string) (GoogleTokenClaims, error) {
	if verifier == nil {
		return GoogleTokenClaims{}, errors.New("google token verifier is not configured")
	}
	idToken, err := verifier.Verify(ctx, rawToken)
	if err != nil {
		return GoogleTokenClaims{}, err
	}
	claims := GoogleTokenClaims{}
	if err := idToken.Claims(&claims); err != nil {
		return GoogleTokenClaims{}, err
	}
	if err := ValidateGoogleTokenClaims(claims, clientID, issuer); err != nil {
		return GoogleTokenClaims{}, err
	}
	return claims, nil
}

func ValidateGoogleTokenClaims(claims GoogleTokenClaims, expectedClientID string, expectedIssuer string) error {
	if claims.Subject == "" {
		return errors.New("google subject is required")
	}
	if claims.Email == "" {
		return errors.New("google email is required")
	}
	if !claims.EmailVerified {
		return errors.New("google email must be verified")
	}
	if expectedIssuer == "" {
		return errors.New("google issuer is not configured")
	}
	if claims.Issuer != expectedIssuer {
		return fmt.Errorf("unexpected google issuer: %s", claims.Issuer)
	}
	if expectedClientID == "" {
		return errors.New("google client id is not configured")
	}
	if claims.Audience != expectedClientID {
		return fmt.Errorf("unexpected google audience: %s", claims.Audience)
	}
	if time.Now().Unix() >= claims.Expiry {
		return errors.New("google token is expired")
	}
	return nil
}

func VerifyGoogleIDToken(ctx context.Context, rawToken string, clientID string, issuer string) (GoogleTokenClaims, error) {
	verifier, err := NewGoogleTokenVerifier(ctx, clientID, issuer)
	if err != nil {
		return GoogleTokenClaims{}, err
	}
	return VerifyGoogleIDTokenWithVerifier(ctx, verifier, rawToken, clientID, issuer)
}
