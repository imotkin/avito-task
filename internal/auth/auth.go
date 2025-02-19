package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

var (
	ErrTokenNotFound   = errors.New("jwt token not found")
	ErrInvalidUserType = errors.New("failed to cast user id to string")
	ErrInvalidUUID     = errors.New("failed to parse user id")
)

type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token"`
}

type Authorizer interface {
	CreateToken(id uuid.UUID, name string) (*Token, error)
	ParseToken(r *http.Request) (uuid.UUID, error)
}

type Service struct {
	auth       *jwtauth.JWTAuth
	expiration time.Duration
}

func NewService(auth *jwtauth.JWTAuth, expiration time.Duration) *Service {
	return &Service{auth: auth, expiration: expiration}
}

func (s *Service) Auth() *jwtauth.JWTAuth {
	return s.auth
}

func (s *Service) CreateToken(id uuid.UUID, name string) (*Token, error) {
	_, token, err := s.auth.Encode(map[string]any{
		"user_id":  id,
		"username": name,
		"exp":      time.Now().Add(s.expiration).Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &Token{token}, nil
}

func (s *Service) ParseToken(r *http.Request) (uuid.UUID, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil || len(claims) == 0 {
		return uuid.Nil, ErrTokenNotFound
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, ErrInvalidUserType
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, ErrInvalidUUID
	}

	return id, nil
}
