package auth

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateToken(t *testing.T) {
	id := uuid.New()

	tokenValue, err := CreateToken(id, "test-user")
	assert.NoError(t, err)

	assert.NotNil(t, tokenValue)

	token, err := TokenJWT.Decode(tokenValue.Token)
	assert.NoError(t, err)

	decodedID, ok := token.Get("user_id")
	assert.True(t, ok)

	decodedName, ok := token.Get("username")
	assert.True(t, ok)

	assert.Equal(t, id.String(), decodedID.(string))
	assert.Equal(t, "test-user", decodedName.(string))
}
