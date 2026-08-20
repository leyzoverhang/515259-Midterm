package user

import (
	"wongnok/internal/convutil"
)

type CreateUserRequest struct {
	Name  *string `json:"name"`
	Email string  `json:"email"`
}

func (req CreateUserRequest) ToUser() User {
	return User{
		Email: req.Email,
		Name:  req.Name,
	}
}

type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

func NewUserResponse(user User) UserResponse {
	return UserResponse{
		ID:    user.ID.String(),
		Name:  convutil.ToSafeValue(user.Name),
		Email: user.Email,
	}
}
