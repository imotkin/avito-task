package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/imotkin/avito-task/internal/auth"
	"github.com/imotkin/avito-task/internal/shop"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testDBUser     = "postgres"
	testDBPassword = "password"
	testDBName     = "test_shop"
)

func setupTestContainers(t *testing.T) (string, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     testDBUser,
			"POSTGRES_PASSWORD": testDBPassword,
			"POSTGRES_DB":       testDBName,
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	dbContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)

	dbHost, err := dbContainer.Host(ctx)
	assert.NoError(t, err)
	dbPort, err := dbContainer.MappedPort(ctx, "5432")
	assert.NoError(t, err)

	dbConnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		testDBUser, testDBPassword, dbHost, dbPort.Port(), testDBName,
	)

	db, err := sql.Open("postgres", dbConnStr)
	assert.NoError(t, err)
	defer db.Close()

	err = db.Ping()
	assert.NoError(t, err)

	appReq := testcontainers.ContainerRequest{
		Image:        "avito-shop",
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"DATABASE_USER":     testDBUser,
			"DATABASE_PASSWORD": testDBPassword,
			"DATABASE_HOST":     dbHost,
			"DATABASE_PORT":     dbPort.Port(),
			"DATABASE_NAME":     testDBName,
		},
		WaitingFor: wait.ForListeningPort("8080/tcp"),
	}
	appContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: appReq,
		Started:          true,
	})
	assert.NoError(t, err)

	appHost, err := appContainer.Host(ctx)
	assert.NoError(t, err)
	appPort, err := appContainer.MappedPort(ctx, "8080")
	assert.NoError(t, err)

	baseURL := fmt.Sprintf("http://%s:%s", appHost, appPort.Port())

	return baseURL, func() {
		dbContainer.Terminate(ctx)
		appContainer.Terminate(ctx)
	}
}

func TestIntegration_BuyProduct(t *testing.T) {
	baseURL, cleanup := setupTestContainers(t)
	defer cleanup()

	time.Sleep(2 * time.Second)

	tokenJSON := createUser(t, baseURL)

	var token auth.Token

	err := json.NewDecoder(bytes.NewBufferString(tokenJSON)).Decode(&token)
	assert.Nil(t, err)

	buyProduct(t, baseURL, token.Token, "t-shirt")

	user := getUser(t, baseURL, token.Token)

	assert.Equal(t, 920, user.Coins)
	assert.Contains(t, user.Inventory, "t-shirt")
}

func createUser(t *testing.T, URL string) string {
	endpoint := URL + "/api/auth/"

	body := bytes.NewBufferString(
		`{"username": "ilya", "password": "secret"}`,
	)

	req, err := http.NewRequest(http.MethodPost, endpoint, body)
	assert.Nil(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	return string(b)
}

func buyProduct(t *testing.T, URL, token, product string) {
	endpoint := URL + "/api/buy/" + product

	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	assert.Nil(t, err)

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func getUser(t *testing.T, URL, token string) (user *shop.User) {
	endpoint := URL + "/api/info"

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	assert.Nil(t, err)

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	assert.Nil(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	json.NewDecoder(resp.Body).Decode(user)
	assert.Nil(t, err)

	return
}
