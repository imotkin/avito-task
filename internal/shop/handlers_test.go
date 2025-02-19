package shop

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/imotkin/avito-task/internal/auth"
	me "github.com/imotkin/avito-task/internal/myerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (r *MockRepository) HasUser(ctx context.Context, data auth.LoginData) (uuid.UUID, bool, error) {
	args := r.Called(ctx, data)
	return args.Get(0).(uuid.UUID), args.Bool(1), args.Error(2)
}

func (r *MockRepository) AddUser(ctx context.Context, data auth.LoginData) (uuid.UUID, error) {
	args := r.Called(ctx, data)
	return args.Get(0).(uuid.UUID), args.Error(1)
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

type MockAuth struct {
	mock.Mock
}

func (m *MockAuth) CreateToken(id uuid.UUID, name string) (*auth.Token, error) {
	args := m.Called(id, name)
	return args.Get(0).(*auth.Token), args.Error(1)
}

func (m *MockAuth) ParseToken(r *http.Request) (uuid.UUID, error) {
	args := m.Called(r)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func NewTestRequest(
	method string, target string, body io.Reader, headers map[string]string,
) *http.Request {
	req := httptest.NewRequest(method, target, body)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req
}

func TestAuthorize(t *testing.T) {
	mockRepo := new(MockRepository)
	mockAuth := new(MockAuth)

	service := NewService(mockRepo, mockAuth, nil)

	tests := []struct {
		name     string
		request  string
		status   int
		response string
		fn       func()
	}{
		{
			name:     "Invalid JSON in request",
			request:  `{"1}`,
			response: `{"errors":"invalid JSON"}`,
			status:   http.StatusBadRequest,
		},
		{
			name:     "Empty username value",
			request:  `{"password":"test"}`,
			response: `{"errors":"empty username"}`,
			status:   http.StatusBadRequest,
		},
		{
			name:     "Empty password value",
			request:  `{"username":"test"}`,
			response: `{"errors":"empty password"}`,
			status:   http.StatusBadRequest,
		},
		{
			name:     "Create a new user",
			request:  `{"username": "ilya", "password": "qwerty"}`,
			response: `{"token":"mocked-token"}`,
			status:   http.StatusOK,
			fn: func() {
				data := auth.LoginData{
					Username: "ilya", Password: "qwerty",
				}

				id := uuid.New()

				mockRepo.On("HasUser", mock.Anything, data).
					Return(uuid.Nil, false, nil).
					Once()

				mockRepo.On("AddUser", mock.Anything, data).
					Return(id, nil).
					Once()

				mockAuth.On("CreateToken", id, "ilya").
					Return(&auth.Token{Token: "mocked-token"}, nil).
					Once()
			},
		},
		{
			name:     "Invalid user password",
			request:  `{"username": "ilya", "password": "qwerty123"}`,
			response: `{"errors":"invalid user password"}`,
			status:   http.StatusBadRequest,
			fn: func() {
				data := auth.LoginData{
					Username: "ilya", Password: "qwerty123",
				}

				id := uuid.New()

				mockRepo.On("HasUser", mock.Anything, data).
					Return(id, true, me.ErrInvalidPassword).
					Once()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fn != nil {
				tt.fn()
			}

			req := httptest.NewRequest(http.MethodPost, "/api/auth", bytes.NewBufferString(tt.request))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			service.Authorize(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.status, resp.StatusCode)

			assert.JSONEq(t, tt.response, rec.Body.String())

			mockRepo.AssertExpectations(t)
			mockAuth.AssertExpectations(t)
		})
	}
}

func TestUserInfo(t *testing.T) {
	mockRepo := new(MockRepository)
	mockAuth := new(MockAuth)

	service := NewService(mockRepo, mockAuth, nil)

	tests := []struct {
		name     string
		request  *http.Request
		status   int
		response string
		fn       func()
	}{
		{
			name: "Invalid JWT token",
			request: NewTestRequest(http.MethodGet, "/api/info", nil, map[string]string{
				"Authorization": "Bearer invalid-token",
			}),
			status:   http.StatusBadRequest,
			response: `{"errors": "invalid JWT token"}`,
			fn: func() {
				mockAuth.On("ParseToken", mock.Anything).
					Return(uuid.Nil, auth.ErrTokenNotFound).
					Once()
			},
		},
		// TODO: add token creation
		// {
		// 	name: "Invalid user id",
		// 	request: NewTestRequest(http.MethodGet, "/api/info", nil, map[string]string{
		// 		"Authorization": "Bearer",
		// 	}),
		// 	status:   http.StatusInternalServerError,
		// 	response: `{"errors": "failed to get user info"}`,
		// 	fn: func() {
		// 		id := uuid.New()

		// 		mockAuth.On("ParseToken", mock.Anything).
		// 			Return(id, nil).
		// 			Once()

		// 		mockRepo.On("UserInfo", mock.Anything, id).
		// 			Return(nil, errors.New("")).
		// 			Once()
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.fn != nil {
				tt.fn()
			}

			rec := httptest.NewRecorder()
			service.UserInfo(rec, tt.request)

			resp := rec.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.status, resp.StatusCode)

			assert.JSONEq(t, tt.response, rec.Body.String())

			mockRepo.AssertExpectations(t)
			mockAuth.AssertExpectations(t)
		})
	}
}
