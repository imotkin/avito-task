package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"

	"github.com/imotkin/avito-task/internal/auth"
	"github.com/imotkin/avito-task/internal/config"
	"github.com/imotkin/avito-task/internal/database"
	"github.com/imotkin/avito-task/internal/migrations"
	"github.com/imotkin/avito-task/internal/shop"
	"github.com/joho/godotenv"
)

func parseLevel(name string) (slog.Level, error) {
	switch strings.ToLower(name) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, errors.New("invalid logging level")
	}
}

func main() {
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Println("Failed to load .env file, default settings were set")
	}

	cfg := config.Load()

	level, err := parseLevel(cfg.Logging)
	if err != nil {
		log.Println("Failed to parse logging level, default level was set (info)")
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	err = migrations.Up(db.Connection())
	if err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	authService := auth.NewService(
		jwtauth.New("HS256", []byte("secret"), nil), (time.Hour * 24),
	)

	shopService := shop.NewService(db, authService, logger)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(authService.Auth()))

		r.Use(jwtauth.Authenticator(authService.Auth()))

		r.Get("/api/info", shopService.UserInfo)
		r.Get("/api/buy/{item}", shopService.BuyProduct)
		r.Get("/api/sendCoin", shopService.SendCoin)
	})

	r.Post("/api/auth", shopService.Authorize)

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	go func() {
		log.Printf("Got an interrupt signal: %v\n", <-sigs)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			log.Printf("Failed to shutdown HTTP server: %v\n", err)
		}

		log.Println("Server HTTP was closed")

		done <- true
	}()

	go func() {
		log.Printf("Started HTTP server at http://localhost%s\n", server.Addr)

		err = server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	<-done
}
