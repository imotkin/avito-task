package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"slices"
	"testing"

	"github.com/imotkin/avito-task/internal/auth"
	"github.com/imotkin/avito-task/internal/shop"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationBuyProduct(t *testing.T) {
	cleanup := setupTestContainers(t)
	defer cleanup()

	token := createUser(t)

	buyProduct(t, token, "t-shirt")

	user := getUser(t, token)

	require.Equal(t, uint64(920), user.Coins)

	var items []shop.Item

	err := json.Unmarshal(user.Inventory, &items)
	require.NoError(t, err)

	require.True(t, slices.ContainsFunc(items, func(item shop.Item) bool {
		return item.Type == "t-shirt" && item.Quantity == 1
	}))
}

func createUser(t *testing.T) string {
	endpoint := baseURL + "/api/auth"

	body := bytes.NewBufferString(
		`{"username": "ilya", "password": "secret"}`,
	)

	req, err := http.NewRequest(http.MethodPost, endpoint, body)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var token auth.Token

	err = json.NewDecoder(resp.Body).Decode(&token)
	require.NoError(t, err)

	require.NotEmpty(t, token.Token)

	return token.Token
}

func buyProduct(t *testing.T, token, product string) {
	endpoint := baseURL + "/api/buy/" + product

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	require.NoError(t, err)

	req.Header.Add("Authorization", ("Bearer " + token))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func getUser(t *testing.T, token string) *shop.User {
	endpoint := baseURL + "/api/info"

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	require.NoError(t, err)

	req.Header.Add("Authorization", ("Bearer " + token))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var user shop.User

	err = json.NewDecoder(resp.Body).Decode(&user)
	require.NoError(t, err)

	return &user
}
