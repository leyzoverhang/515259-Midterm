package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"wongnok/internal/config"
	"wongnok/internal/user"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Repository interface {
	SaveState(ctx context.Context, state string, ttl time.Duration) error
	ConsumeState(ctx context.Context, state string) error
	SaveTicket(ctx context.Context, ticket string, credential Credential, ttl time.Duration) error
	ConsumeTicket(ctx context.Context, ticket string) (Credential, error)
}

type UserService interface {
	UpsertFromKeycloak(ctx context.Context, kuser user.KeycloakUser) error
}

type service struct {
	repository  Repository
	userService UserService
	keycloak    config.Keycloak
	oauth2      *oauth2.Config
	verifier    *oidc.IDTokenVerifier
	http        *http.Client
}

func NewService(repo Repository, userService UserService, keycloak config.Keycloak, provider *oidc.Provider, verifier *oidc.IDTokenVerifier) *service {
	oauthConf := &oauth2.Config{
		ClientID:     keycloak.ClientID,
		ClientSecret: keycloak.ClientSecret,
		RedirectURL:  keycloak.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &service{
		repository:  repo,
		userService: userService,
		keycloak:    keycloak,
		oauth2:      oauthConf,
		verifier:    verifier,
		http:        &http.Client{Timeout: (10 * time.Second)},
	}
}

func (svc *service) BuildLoginURL(ctx context.Context) (string, error) {
	state, err := generateRandomToken()
	if err != nil {
		return "", fmt.Errorf("generate state: %w", err)
	}

	if err := svc.repository.SaveState(ctx, state, (5 * time.Minute)); err != nil {
		return "", fmt.Errorf("save state: %w", err)
	}

	return svc.oauth2.AuthCodeURL(state), nil
}

func (svc *service) HandleCallback(ctx context.Context, code, state string) (string, error) {
	if err := svc.repository.ConsumeState(ctx, state); err != nil {
		if errors.Is(err, ErrNotFound) {
			return "", ErrInvalidState
		}
	}

	credential, claims, err := svc.exchangeCode(ctx, code)
	if err != nil {
		return "", fmt.Errorf("exchange code: %w", err)
	}

	if err := svc.userService.UpsertFromKeycloak(ctx, user.KeycloakUser{
		UID:               claims.Subject,
		Email:             claims.Email,
		Name:              claims.Name,
		PreferredUsername: claims.PreferredUsername,
	}); err != nil {
		return "", fmt.Errorf("upsert user: %w", err)
	}

	ticket, err := generateRandomToken()
	if err != nil {
		return "", fmt.Errorf("generate ticket: %w", err)
	}

	if err := svc.repository.SaveTicket(ctx, ticket, credential, (30 * time.Second)); err != nil {
		return "", fmt.Errorf("save ticket: %w", err)
	}

	return fmt.Sprintf("%s/auth/callback?ticket=%s", svc.keycloak.FrontendURL, ticket), nil
}

func (svc *service) ExchangeTicket(ctx context.Context, ticket string) (Credential, error) {
	credential, err := svc.repository.ConsumeTicket(ctx, ticket)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Credential{}, ErrInvalidTicket
		}
		return Credential{}, fmt.Errorf("consume ticket: %w", err)
	}

	return credential, nil
}

func (svc *service) Logout(ctx context.Context, refreshToken string) error {
	logoutURL := fmt.Sprintf("%s/protocol/openid-connect/logout", svc.keycloak.RealmURL())

	form := url.Values{
		"client_id":     {svc.keycloak.ClientID},
		"client_secret": {svc.keycloak.ClientSecret},
		"refresh_token": {refreshToken},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, logoutURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("build logout request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := svc.http.Do(req)
	if err != nil {
		return fmt.Errorf("call keycloak logout: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return ErrLogoutFailed
	}

	return nil
}

// Private
func generateRandomToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func (svc *service) exchangeCode(ctx context.Context, code string) (Credential, KeycloakClaims, error) {
	token, err := svc.oauth2.Exchange(ctx, code)
	if err != nil {
		return Credential{}, KeycloakClaims{}, fmt.Errorf("%w: %v", ErrExchangeFailed, err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return Credential{}, KeycloakClaims{}, fmt.Errorf("missing id_token")
	}

	idToken, err := svc.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return Credential{}, KeycloakClaims{}, fmt.Errorf("verify id_token: %w", err)
	}

	var claims KeycloakClaims
	if err := idToken.Claims(&claims); err != nil {
		return Credential{}, KeycloakClaims{}, fmt.Errorf("parse claims: %w", err)
	}

	return Credential{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry,
	}, claims, nil
}
