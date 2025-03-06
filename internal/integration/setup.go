package integration

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"testing"

	"github.com/imotkin/avito-task/internal/config"
	"github.com/imotkin/avito-task/internal/migrations"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	baseURL string
)

func setupTestContainers(t *testing.T) func() {
	config := &config.Config{
		Host:       "postgres",
		Port:       "5432",
		User:       "postgres",
		Password:   "postgres",
		Database:   "shop",
		ServerPort: "8080",
	}

	ctx := context.Background()

	err := exec.Command("docker", "info").Run()
	if err != nil {
		t.Fatalf("Docker isn't started: %v", err)
	}

	network, err := network.New(ctx,
		network.WithDriver("bridge"),
		network.WithAttachable(),
	)
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
		Hostname:   "postgres",
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

	appReq := testcontainers.ContainerRequest{
		Image:        "avito-task-avito-shop-service",
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"DATABASE_USER":     config.User,
			"DATABASE_PASSWORD": config.Password,
			"DATABASE_HOST":     "postgres", // nit: service and db containers are in the same newtork, so it's necessary to use hostname instead of external host and default port for PostgreSQL: 5432. Another alternative to use a detached container for database migrations, but that seems too wasteful
			"DATABASE_PORT":     "5432",
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

	logs, _ := appContainer.Logs(ctx)
	io.Copy(os.Stdout, logs)

	require.NoError(t, err)

	appHost, err := appContainer.Host(ctx)
	require.NoError(t, err)

	appPort, err := appContainer.MappedPort(ctx, "8080")
	require.NoError(t, err)

	baseURL = fmt.Sprintf("http://%s:%s", appHost, appPort.Port())

	return func() {
		err := migrations.Down(db, "../../migrations")
		if err != nil {
			log.Printf("Failed to run down migrations: %v", err)
		}

		err = dbContainer.Terminate(ctx)
		if err != nil {
			log.Printf("Failed to stop DB container: %v", err)
		}

		err = appContainer.Terminate(ctx)
		if err != nil {
			log.Printf("Failed to stop app container: %v", err)
		}

		err = network.Remove(ctx)
		if err != nil {
			log.Printf("Failed to remove network: %s", err)
		}
	}
}
