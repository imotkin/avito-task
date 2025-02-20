package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/imotkin/avito-task/internal/auth"
	"github.com/imotkin/avito-task/internal/config"
	"github.com/imotkin/avito-task/internal/migrations"
	"github.com/imotkin/avito-task/internal/shop"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestContainers(t *testing.T) (string, func()) {
	config := &config.Config{
		User:       "postgres",
		Password:   "postgres",
		Database:   "shop",
		ServerPort: "8080",
	}

	ctx := context.Background()

	network, err := network.New(ctx)
	require.NoError(t, err)

	req := testcontainers.ContainerRequest{
		Image:        "postgres:17",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     config.User,
			"POSTGRES_PASSWORD": config.Password,
			"POSTGRES_DB":       config.Database,
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
		Networks:   []string{network.Name},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.NetworkMode = testcontainers.Bridge
		},
	}

	dbContainer, err := testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		},
	)

	require.NoError(t, err)

	dbHost, err := dbContainer.Host(ctx)
	require.NoError(t, err)

	dbPort, err := dbContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	config.Host = dbHost
	config.Port = dbPort.Port()

	db, err := sql.Open("postgres", config.DatabaseURL())
	require.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	require.NoError(t, err)

	err = migrations.Up(db, "../../migrations")
	require.NoError(t, err)

	fmt.Printf("%+v\n", config)

	appReq := testcontainers.ContainerRequest{
		Image:        "avito-task-avito-shop-service",
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"DATABASE_USER":     config.User,
			"DATABASE_PASSWORD": config.Password,
			"DATABASE_HOST":     "host.docker.internal",
			"DATABASE_PORT":     config.Port,
			"DATABASE_NAME":     config.Database,
			"SERVER_PORT":       config.ServerPort,
		},
		WaitingFor: wait.ForLog("Started HTTP server"),
		Networks:   []string{network.Name},
	}

	appContainer, err := testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: appReq,
			Started:          true,
		},
	)

	require.NoError(t, err)

	logs, _ := appContainer.Logs(ctx)
	io.Copy(os.Stdout, logs)

	appHost, err := appContainer.Host(ctx)
	require.NoError(t, err)

	appPort, err := appContainer.MappedPort(ctx, "8080")
	require.NoError(t, err)

	baseURL := fmt.Sprintf("http://%s:%s", appHost, appPort.Port())

	return baseURL, func() {
		migrations.Down(db, "../../migrations")
		dbContainer.Terminate(ctx)
		appContainer.Terminate(ctx)
	}
}

func TestIntegrationBuyProduct(t *testing.T) {
	baseURL, cleanup := setupTestContainers(t)
	defer cleanup()

	token := createUser(t, baseURL)

	buyProduct(t, baseURL, token, "t-shirt")

	user := getUser(t, baseURL, token)

	require.Equal(t, uint64(920), user.Coins)

	var items []shop.Item

	err := json.Unmarshal(user.Inventory, &items)
	require.NoError(t, err)

	require.True(t, slices.ContainsFunc(items, func(item shop.Item) bool {
		return item.Type == "t-shirt" && item.Quantity == 1
	}))
}

func TestIntegrationSendCoin(t *testing.T) {
	baseURL, cleanup := setupTestContainers(t)
	defer cleanup()

	tokenAuth := jwtauth.New("HS256", []byte("secret"), nil)

	tokenSender := createUser(t, baseURL)
	tokenReciever := createUser(t, baseURL)

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

	sendCoin(t, baseURL, tokenSender, sentTransfer)

	user := getUser(t, baseURL, tokenSender)

	var sent []shop.Transfer

	err = json.Unmarshal(user.CoinHistory.Sent, &sent)
	require.NoError(t, err)

	require.True(t, slices.ContainsFunc(sent, func(t shop.Transfer) bool {
		return reflect.DeepEqual(t, sentTransfer)
	}))
}

func createUser(t *testing.T, URL string) string {
	endpoint := URL + "/api/auth"

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

func sendCoin(t *testing.T, URL, token string, transfer shop.Transfer) {
	endpoint := URL + "/api/sendCoin"

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

func buyProduct(t *testing.T, URL, token, product string) {
	endpoint := URL + "/api/buy/" + product

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	require.NoError(t, err)

	req.Header.Add("Authorization", ("Bearer " + token))

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func getUser(t *testing.T, URL, token string) *shop.User {
	endpoint := URL + "/api/info"

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
