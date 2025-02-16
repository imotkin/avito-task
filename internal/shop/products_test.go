package shop

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsProduct(t *testing.T) {
	tests := []struct {
		product string
		ok      bool
	}{
		{"t-shirt", true},
		{"_pen_", false},
		{"", false},
		{"umbrella", true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, IsProduct(tt.product), tt.ok)
		})
	}
}
