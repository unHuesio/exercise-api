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

// NewGoogleTokenVerifier builds a verifier that accepts ID tokens issued for
// any of the given clientIDs (e.g. separate web and iOS OAuth clients).
// Audience checking is done manually in ValidateGoogleTokenClaims because
// go-oidc's Config.ClientID only supports a single expected audience.
func NewGoogleTokenVerifier(ctx context.Context, clientIDs []string, issuer string) (*oidc.IDTokenVerifier, error) {
	if len(clientIDs) == 0 {
		return nil, errors.New("google client id is not configured")
	}
	if issuer == "" {
		return nil, errors.New("google issuer is not configured")
	}
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}
	return provider.Verifier(&oidc.Config{SkipClientIDCheck: true}), nil
}

func VerifyGoogleIDTokenWithVerifier(ctx context.Context, verifier *oidc.IDTokenVerifier, rawToken string, clientIDs []string, issuer string) (GoogleTokenClaims, error) {
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
	if err := ValidateGoogleTokenClaims(claims, clientIDs, issuer); err != nil {
		return GoogleTokenClaims{}, err
	}
	return claims, nil
}

func ValidateGoogleTokenClaims(claims GoogleTokenClaims, expectedClientIDs []string, expectedIssuer string) error {
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
	if len(expectedClientIDs) == 0 {
		return errors.New("google client id is not configured")
	}
	audienceAllowed := false
	for _, id := range expectedClientIDs {
		if claims.Audience == id {
			audienceAllowed = true
			break
		}
	}
	if !audienceAllowed {
		return fmt.Errorf("unexpected google audience: %s", claims.Audience)
	}
	if time.Now().Unix() >= claims.Expiry {
		return errors.New("google token is expired")
	}
	return nil
}

func VerifyGoogleIDToken(ctx context.Context, rawToken string, clientIDs []string, issuer string) (GoogleTokenClaims, error) {
	verifier, err := NewGoogleTokenVerifier(ctx, clientIDs, issuer)
	if err != nil {
		return GoogleTokenClaims{}, err
	}
	return VerifyGoogleIDTokenWithVerifier(ctx, verifier, rawToken, clientIDs, issuer)
}
