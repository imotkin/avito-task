package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateToken(t *testing.T) {
	authService := NewService(
		jwtauth.New("HS256", []byte("secret"), nil), (time.Hour * 24),
	)

	id := uuid.New()

	tokenValue, err := authService.CreateToken(id, "test-user")
	assert.NoError(t, err)

	assert.NotNil(t, tokenValue)

	token, err := authService.Auth().Decode(tokenValue.Token)
	assert.NoError(t, err)

	decodedID, ok := token.Get("user_id")
	assert.True(t, ok)

	decodedName, ok := token.Get("username")
	assert.True(t, ok)

	assert.Equal(t, id.String(), decodedID.(string))
	assert.Equal(t, "test-user", decodedName.(string))
}

func TestParseToken(t *testing.T) {
	authService := NewService(
		jwtauth.New("HS256", []byte("secret"), nil), (time.Hour * 24),
	)

	tests := []struct {
		name string
		fn   func(t *testing.T) *http.Request
		err  error
	}{
		{
			name: "Valid token",
			fn: func(t *testing.T) *http.Request {
				token, tokenString, err := authService.Auth().Encode(map[string]any{
					"user_id":  uuid.New().String(),
					"username": "ilya",
					"exp":      time.Now().Add(time.Hour * 24).Unix(),
				})
				assert.NoError(t, err)

				ctx := jwtauth.NewContext(context.Background(), token, nil)

				req := httptest.NewRequestWithContext(ctx, "GET", "/", nil)
				req.Header.Set("Authorization", "Bearer "+tokenString)

				return req
			},
			err: nil,
		},
		{
			name: "Empty context",
			fn: func(t *testing.T) *http.Request {
				return httptest.NewRequest("GET", "/", nil)
			},
			err: ErrTokenNotFound,
		},
		{
			name: "Invalid user id type",
			fn: func(t *testing.T) *http.Request {
				token, tokenString, err := authService.Auth().Encode(map[string]any{
					"user_id":  12345,
					"username": "ilya",
					"exp":      time.Now().Add(time.Hour * 24).Unix(),
				})
				assert.NoError(t, err)

				ctx := jwtauth.NewContext(context.Background(), token, nil)

				req := httptest.NewRequestWithContext(ctx, "GET", "/", nil)
				req.Header.Set("Authorization", "Bearer "+tokenString)

				return req
			},
			err: ErrInvalidUserType,
		},
		{
			name: "Invalid uuid value",
			fn: func(t *testing.T) *http.Request {
				token, tokenString, err := authService.Auth().Encode(map[string]any{
					"user_id":  "12345",
					"username": "ilya",
					"exp":      time.Now().Add(time.Hour * 24).Unix(),
				})
				assert.NoError(t, err)

				ctx := jwtauth.NewContext(context.Background(), token, nil)

				req := httptest.NewRequestWithContext(ctx, "GET", "/", nil)
				req.Header.Set("Authorization", "Bearer "+tokenString)

				return req
			},
			err: ErrInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.fn(t)
			_, err := authService.ParseToken(req)
			assert.Equal(t, tt.err, err)
		})
	}
}
