package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"slices"
	"testing"

	"github.com/imotkin/avito-task/internal/shop"

	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestIntegrationSendCoin(t *testing.T) {
	cleanup := setupTestContainers(t)
	defer cleanup()

	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	tokenSender := createUser(t)
	tokenReciever := createUser(t)

	tokenDecoded, err := tokenAuth.Decode(tokenReciever)
	require.NoError(t, err)

	recieverID, ok := tokenDecoded.Get("user_id")
	require.True(t, ok)

	id, ok := recieverID.(string)
	require.True(t, ok)

	uuid, err := uuid.Parse(id)
	require.NoError(t, err)

	sentTransfer := shop.Transfer{
		Receiver: uuid,
		Amount:   100,
	}

	sendCoin(t, tokenSender, sentTransfer)

	user := getUser(t, tokenSender)

	var sent []shop.Transfer

	err = json.Unmarshal(user.CoinHistory.Sent, &sent)
	require.NoError(t, err)

	require.True(t, slices.ContainsFunc(sent, func(t shop.Transfer) bool {
		return reflect.DeepEqual(t, sentTransfer)
	}))
}

func sendCoin(t *testing.T, token string, transfer shop.Transfer) {
	endpoint := baseURL + "/api/sendCoin"

	payload, err := json.Marshal(transfer)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, endpoint, bytes.NewBuffer(payload))
	require.NoError(t, err)

	req.Header.Add("Authorization", ("Bearer " + token))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}
