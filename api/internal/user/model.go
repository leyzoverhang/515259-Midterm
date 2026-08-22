package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                uuid.UUID `gorm:"primaryKey;default:gen_random_uuid()"`
	Email             string
	Name              *string
	Bio               *string
	UID               string
	PreferredUsername *string
	LastSignedInAt    *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt
}

type KeycloakUser struct {
	UID               string
	Email             string
	Name              string
	PreferredUsername string
}
