package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"wongnok/internal/httputil"
)

type Service interface {
	FindByID(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, user User) (*User, error)
}

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (hdr *handler) GetUser(writer http.ResponseWriter, request *http.Request) {
	uid := request.PathValue("id")

	user, err := hdr.service.FindByID(request.Context(), uid)
	if err != nil {
		writer.Header().Set("Content-Type", "application/json")

		switch {
		case errors.Is(err, ErrUserNotFound):
			httputil.WriteError(writer, http.StatusNotFound, httputil.ErrorResponse{Message: err.Error()})

		case errors.Is(err, ErrInvalidInput):
			httputil.WriteError(writer, http.StatusBadRequest, httputil.ErrorResponse{Message: err.Error()})

		default:
			httputil.WriteError(writer, http.StatusInternalServerError, httputil.ErrorResponse{Message: err.Error()})

		}

		return
	}

	httputil.WriteJSON(writer, http.StatusOK, NewUserResponse(*user))
}

func (hdr *handler) CreateUser(writer http.ResponseWriter, request *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		httputil.WriteError(writer, http.StatusBadRequest, httputil.ErrorResponse{Message: "invalid request body"})
		return
	}

	user, err := hdr.service.Create(request.Context(), req.ToUser())
	if err != nil {
		httputil.WriteError(writer, http.StatusBadRequest, httputil.ErrorResponse{Message: err.Error()})
		return
	}

	httputil.WriteJSON(writer, http.StatusCreated, NewUserResponse(*user))
}
