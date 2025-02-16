package shop

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/imotkin/avito-task/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (r *MockRepository) HasUser(ctx context.Context, data auth.LoginData) (uuid.UUID, bool, error) {
	args := r.Called(ctx, data)
	return uuid.MustParse(args.String(0)), args.Bool(1), args.Error(2)
}

func (r *MockRepository) AddUser(ctx context.Context, data auth.LoginData) (uuid.UUID, error) {
	args := r.Called(ctx, data)
	return uuid.MustParse(args.String(0)), args.Error(1)
}

func (r *MockRepository) HasUserID(ctx context.Context, userID uuid.UUID) (bool, error) {
	args := r.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (r *MockRepository) BuyProduct(ctx context.Context, userID uuid.UUID, item string) error {
	return nil
}

func (r *MockRepository) SendCoin(ctx context.Context, transfer Transfer) error {
	return nil
}

func (r *MockRepository) UserInfo(ctx context.Context, userID uuid.UUID) (*User, error) {
	return nil, nil
}

func TestAuthorize(t *testing.T) {
	service := NewService(new(MockRepository), nil)

	tests := []struct {
		request  string
		status   int
		response string
	}{
		{`{"1}`, http.StatusBadRequest, `{"errors":"invalid JSON"}`},
		{`{"username":""}`, http.StatusBadRequest, `{"errors":"empty username"}`},
		{`{"username":"test"}`, http.StatusBadRequest, `{"errors":"empty password"}`},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewBufferString(tt.request))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			service.Authorize(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.status, resp.StatusCode)
			assert.JSONEq(t, tt.response, rec.Body.String())
		})
	}
}
