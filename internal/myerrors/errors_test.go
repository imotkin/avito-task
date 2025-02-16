package myerrors

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected ErrorMessage
	}{
		{"String value", "error message", ErrorMessage{"error message"}},
		{"Error value", errors.New("invalid JSON"), ErrorMessage{"invalid JSON"}},
		{"Unknown type", 100, UnknownError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := New(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSendError(t *testing.T) {
	tests := []struct {
		name         string
		message      any
		sentCode     int
		expectedJSON string
		expectedCode int
	}{
		{"Internal error", "Internal Server Error", 500, `{"errors":"Internal Server Error"}`, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			r := httptest.NewRecorder()

			Error(r, req, tt.message, tt.sentCode)

			assert.Equal(t, tt.expectedCode, r.Code)
			assert.JSONEq(t, tt.expectedJSON, r.Body.String())
		})
	}
}
