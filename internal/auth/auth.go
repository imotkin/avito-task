package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
)

var (
	TokenJWT *jwtauth.JWTAuth
	TTL      = time.Hour * 24
)

func init() {
	TokenJWT = jwtauth.New("HS256", []byte("secret"), nil)
}

type LoginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token"`
}

func CreateToken(id uuid.UUID, name string) (*Token, error) {
	_, token, err := TokenJWT.Encode(map[string]any{
		"user_id":  id,
		"username": name,
		"exp":      time.Now().Add(TTL).Unix(),
	})
	if err != nil {
		return nil, err
	}

	return &Token{token}, nil
}

func ParseToken(r *http.Request) (uuid.UUID, error) {
	_, claims, err := jwtauth.FromContext(r.Context())
	if err != nil {
		return uuid.Nil, fmt.Errorf("get token from context: %v", err)
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("cast id to string: %v", err)
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("cast id to string: %v", err)
	}

	return id, nil
}
