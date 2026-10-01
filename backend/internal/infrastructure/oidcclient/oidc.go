// Package oidcclient wraps an OIDC provider with PKCE login and ID token verification.
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
	// PreferredUsername is optional in OIDC, so it is often empty.
	PreferredUsername string
}

const stateExpiry = 10 * time.Minute

type Client struct {
	stateRepo repository.OidcStateRepository
	provider  *oidc.Provider
	verifier  *oidc.IDTokenVerifier
	oauth     oauth2.Config
}

// New runs OIDC discovery so a misconfigured issuer fails at startup.
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

// AuthorizationURL starts a PKCE login and returns the redirect URL and state.
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

// Exchange redeems the code for tokens and verifies the ID token.
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

	// Profile and email claims come from userinfo; the ID token only guarantees sub.
	userInfo, err := c.provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
	if err != nil {
		return nil, fmt.Errorf("fetching userinfo failed: %w", err)
	}
	// The userinfo subject must match the ID token's (OIDC spec).
	if userInfo.Subject != idToken.Subject {
		return nil, fmt.Errorf("userinfo subject %q does not match id_token subject %q", userInfo.Subject, idToken.Subject)
	}

	var profile struct {
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
		// Returned thanks to the "profile" scope.
		PreferredUsername string `json:"preferred_username"`
	}
	if err := userInfo.Claims(&profile); err != nil {
		return nil, fmt.Errorf("failed to parse userinfo claims: %w", err)
	}

	return &Identity{
		Subject:       idToken.Subject,
		Email:         userInfo.Email,
		EmailVerified: userInfo.EmailVerified,
		FirstName:     profile.GivenName,
		LastName:      profile.FamilyName,

		PreferredUsername: profile.PreferredUsername,
	}, nil
}

func randomString() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}
