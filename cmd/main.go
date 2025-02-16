package main

import (
	"cmp"
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/imotkin/avito-task/internal/auth"
	"github.com/imotkin/avito-task/internal/config"
	"github.com/imotkin/avito-task/internal/database"
	"github.com/imotkin/avito-task/internal/migrations"
	"github.com/imotkin/avito-task/internal/shop"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
)

var (
	logLevel = flag.String("logger", "error", "Set a necessary logger level: info, debug, warn, error")
	levels   = map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}
)

func main() {
	flag.Parse()

	level := cmp.Or(levels[*logLevel], slog.LevelError)

	logger := slog.New(slog.NewTextHandler(
		os.Stdout, &slog.HandlerOptions{
			Level: level,
		},
	))

	cfg := config.Load()

	db, err := database.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	err = migrations.Up(db.Connection())
	if err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	service := shop.NewService(db, logger)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(auth.TokenJWT))

		r.Use(jwtauth.Authenticator(auth.TokenJWT))

		r.Get("/api/info", service.UserInfo)
		r.Get("/api/buy/{item}", service.BuyProduct)
		r.Get("/api/sendCoin", service.SendCoin)
	})

	r.Post("/api/auth", service.Authorize)

	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: r,
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	go func() {
		log.Printf("Got signal: %v\n", <-sigs)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := server.Shutdown(ctx)
		if err != nil {
			log.Printf("Failed to shutdown HTTP server: %v", err)
		}

		log.Println("Server HTTP was closed")

		done <- true
	}()

	go func() {
		log.Printf("Started HTTP server at http://localhost%s", server.Addr)

		err = server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	<-done
}
