package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"evcc-cloud/backend/internal/api"
	"evcc-cloud/backend/internal/storage"

	"github.com/getsentry/sentry-go"
)

func main() {
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:         dsn,
			Environment: os.Getenv("APP_ENV"),
			BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
				event.User = sentry.User{}
				return event
			},
		}); err != nil {
			log.Printf("sentry init failed: %v", err)
		}
		defer sentry.Flush(2 * time.Second)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://evcc:evcc@localhost:5432/evcc_hub?sslmode=disable"
	}

	env := os.Getenv("ENV")
	devMode := env == "development"
	corsOrigin := os.Getenv("CORS_ORIGIN")
	mqttBrokerAddr := os.Getenv("MQTT_BROKER_ADDR")

	db, err := storage.Open(databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := api.NewRouter(db, api.Config{
		JWTSecret:      jwtSecret,
		DevMode:        devMode,
		CORSOrigin:     corsOrigin,
		MQTTBrokerAddr: mqttBrokerAddr,
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("evcc-cloud backend starting on :%s (dev=%v)", port, devMode)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
