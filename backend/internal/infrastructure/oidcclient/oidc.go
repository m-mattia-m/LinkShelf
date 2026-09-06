// Package oidcclient wraps a single, generically-configured OIDC provider:
// discovery, a PKCE-protected authorization-code flow, and ID token
// verification. It intentionally has no knowledge of any specific provider
// (Google, Okta, Zitadel, ...) — whatever is configured under
// authentication.oidc is used as-is.
package oidcclient

import (
	"backend/internal/config"
	"backend/internal/infrastructure/repository"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Identity is the subset of ID token claims we care about.
type Identity struct {
	Subject       string
	Email         string
	EmailVerified bool
	FirstName     string
	LastName      string
}

const stateExpiry = 10 * time.Minute

type Client struct {
	stateRepo repository.OidcStateRepository
	provider  *oidc.Provider
	verifier  *oidc.IDTokenVerifier
	oauth     oauth2.Config
}

// New performs OIDC discovery against the configured issuer. It's called
// once at startup when authentication.type is OIDC, so a misconfigured
// issuer fails fast instead of on the first login attempt. stateRepo persists
// PKCE state/verifier pairs so any backend replica can complete a login
// started on another one - no in-memory or sticky-session requirement.
func New(ctx context.Context, stateRepo repository.OidcStateRepository) (*Client, error) {
	issuer := config.String("authentication.oidc.issuer")
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery against %q failed: %w", issuer, err)
	}

	clientId := config.String("authentication.oidc.clientId")

	return &Client{
		stateRepo: stateRepo,
		provider:  provider,
		verifier:  provider.Verifier(&oidc.Config{ClientID: clientId}),
		oauth: oauth2.Config{
			ClientID:     clientId,
			ClientSecret: config.String("authentication.oidc.clientSecret"),
			RedirectURL:  config.String("authentication.oidc.redirectUrl"),
			Endpoint:     provider.Endpoint(),
			Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
		},
	}, nil
}

// AuthorizationURL starts a new PKCE-protected login attempt and returns the
// URL the frontend should redirect the browser to, along with the state it
// must send back on POST /v1/auth/oidc/callback.
func (c *Client) AuthorizationURL() (authURL, state string, err error) {
	state, err = randomString()
	if err != nil {
		return "", "", err
	}
	verifier := oauth2.GenerateVerifier()

	if err := c.stateRepo.Create(state, verifier, time.Now().Add(stateExpiry)); err != nil {
		return "", "", err
	}

	return c.oauth.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier)), state, nil
}

// Exchange completes a login attempt: it exchanges the authorization code for
// tokens (using the matching PKCE verifier for that state) and verifies the
// returned ID token.
func (c *Client) Exchange(ctx context.Context, code, state string) (*Identity, error) {
	pending, err := c.stateRepo.GetByState(state)
	if err != nil {
		return nil, err
	}
	if pending == nil || pending.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("unknown or expired state")
	}

	// Single-use: remove immediately so the state/verifier can't be replayed.
	if err := c.stateRepo.DeleteByState(state); err != nil {
		return nil, err
	}

	token, err := c.oauth.Exchange(ctx, code, oauth2.VerifierOption(pending.CodeVerifier))
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	rawIdToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, fmt.Errorf("provider did not return an id_token")
	}

	idToken, err := c.verifier.Verify(ctx, rawIdToken)
	if err != nil {
		return nil, fmt.Errorf("id_token verification failed: %w", err)
	}

	var claims struct {
		Subject       string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse id_token claims: %w", err)
	}

	return &Identity{
		Subject:       claims.Subject,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		FirstName:     claims.GivenName,
		LastName:      claims.FamilyName,
	}, nil
}

func randomString() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
