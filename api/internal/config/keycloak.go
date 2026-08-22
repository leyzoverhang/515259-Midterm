package config

import "fmt"

type Keycloak struct {
	BaseURL      string `env:"KEYCLOAK_BASE_URL" envDefault:"http://localhost:8081" validate:"required"`
	Realm        string `env:"KEYCLOAK_REALM" envDefault:"pea" validate:"required"`
	ClientID     string `env:"KEYCLOAK_CLIENT_ID" envDefault:"wongnok" validate:"required"`
	ClientSecret string `env:"KEYCLOAK_CLIENT_SECRET" validate:"required"`
	RedirectURL  string `env:"KEYCLOAK_REDIRECT_URL" envDefault:"http://localhost:8080/api/v1/auth/callback" validate:"required"`
	FrontendURL  string `env:"FRONTEND_URL" envDefault:"http://localhost:3000" validate:"required"`
}

func (keycloak Keycloak) RealmURL() string {
	return fmt.Sprintf("%s/realms/%s", keycloak.BaseURL, keycloak.Realm)
}

func (keycloak Keycloak) Validate() error {
	return validate.Struct(keycloak)
}
