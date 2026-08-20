package user

import (
	"wongnok/internal/convutil"
)

type CreateUserRequest struct {
	Name  *string `json:"name" example:"สมชาย ใจดี"`
	Email string  `json:"email" binding:"required,email" example:"somchai@pea.co.th"`
}

func (req CreateUserRequest) ToUser() User {
	return User{
		Email: req.Email,
		Name:  req.Name,
	}
}

type UserResponse struct {
	ID    string `json:"id" example:"eb238dfe-9d49-4345-9cc9-350c50b429e6"`
	Name  string `json:"name,omitempty" example:"สมชาย ใจดี"`
	Email string `json:"email" example:"somchai@pea.co.th"`
}

func NewUserResponse(user User) UserResponse {
	return UserResponse{
		ID:    user.ID.String(),
		Name:  convutil.ToSafeValue(user.Name),
		Email: user.Email,
	}
}
