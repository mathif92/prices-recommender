package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	serpapiclient "github.com/mathif92/prices-recommender/pkg/client/serpapi"
	"github.com/mathif92/prices-recommender/pkg/collector"
	serpapicollector "github.com/mathif92/prices-recommender/pkg/collector/serpapi"
	"github.com/mathif92/prices-recommender/pkg/job"
	"github.com/mathif92/prices-recommender/pkg/recommendations"
	"github.com/mathif92/prices-recommender/pkg/repositories"
	"github.com/mathif92/prices-recommender/pkg/scheduler"

	"github.com/mathif92/prices-recommender/pkg/api"

	"github.com/mathif92/prices-recommender/frontend"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	if err := godotenv.Load(); err != nil {
		log.Warnf("no .env file found, using defaults: %v", err)
	}

	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getEnv("DB_USER", "prices"),
		getEnv("DB_PASSWORD", "prices"),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_NAME", "prices_recommender"),
	)

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Info("connected to database")

	jwtSecret := getEnv("JWT_SECRET", "change-me-in-production")

	apiClient := serpapiclient.NewClient(
		getEnv("SERP_API_BASE_URL", "https://serpapi.com/search"),
		getEnv("SERP_API_KEY", ""),
	)

	dataRepo := repositories.NewDataRepository(db, db)
	serpapiCollector := serpapicollector.NewCollector(apiClient, dataRepo)
	mainCollector := collector.NewCollector(log, serpapiCollector)
	notifier := recommendations.NewNotifier(log, recommendations.Config{
		SMTPServer: getEnv("SMTP_SERVER", ""),
		SMTPPort:   getEnv("SMTP_PORT", "587"),
		SMTPUser:   getEnv("SMTP_USER", ""),
		SMTPPass:   getEnv("SMTP_PASS", ""),
		SMTPFrom:   getEnv("SMTP_FROM", "prices-recommender@localhost"),
	})

	jobCollector := job.NewCollector(log, db, mainCollector, notifier)

	sched := scheduler.NewScheduler(log, dataRepo, jobCollector)

	apiHandler := api.NewHandler(api.Config{
		Log:            log,
		Repo:           dataRepo,
		Collector:      jobCollector,
		Scheduler:      sched,
		JWTSecret:      jwtSecret,
		GoogleClientID: getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleSecret:   getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirect: fmt.Sprintf("http://localhost:%s/api/auth/google/callback", getEnv("HTTP_PORT", "8080")),
		BaseURL:        fmt.Sprintf("http://localhost:%s", getEnv("HTTP_PORT", "8080")),
	})

	mux := http.NewServeMux()
	mux.Handle("/api/", apiHandler)
	mux.Handle("/", frontend.Handler())

	port := getEnv("HTTP_PORT", "8080")
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	if err := sched.Start(context.Background()); err != nil {
		log.Warnf("failed to start scheduler: %v", err)
	}

	go func() {
		log.Infof("http server listening on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down...")
	sched.Stop()
	server.Shutdown(context.Background())
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
