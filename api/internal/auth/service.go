package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
	"wongnok/internal/config"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Repository interface {
	SaveState(ctx context.Context, state string, ttl time.Duration) error
}

type service struct {
	repository Repository
	keycloak   config.Keycloak
	oauth2     *oauth2.Config
}

func NewService(repo Repository, keycloak config.Keycloak, provider *oidc.Provider) *service {
	oauthConf := &oauth2.Config{
		ClientID:     keycloak.ClientID,
		ClientSecret: keycloak.ClientSecret,
		RedirectURL:  keycloak.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	return &service{
		repository: repo,
		keycloak:   keycloak,
		oauth2:     oauthConf,
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

// Private
func generateRandomToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}
