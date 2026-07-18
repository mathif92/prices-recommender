package testhelpers

import (
	"context"
	"embed"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type TestDB struct {
	DB        *sqlx.DB
	Container *postgres.PostgresContainer
}

func StartTestDB() (*sqlx.DB, func(), error) {
	ctx := context.Background()

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:17-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategyAndDeadline(180*time.Second,
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(1),
			wait.ForListeningPort("5432/tcp"),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("start container: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("connection string: %w", err)
	}

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("connect: %w", err)
	}

	if err := runMigrations(db); err != nil {
		db.Close()
		container.Terminate(ctx)
		return nil, nil, fmt.Errorf("migrations: %w", err)
	}

	cleanup := func() {
		db.Close()
		container.Terminate(context.Background())
	}

	return db, cleanup, nil
}

func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	ctx := context.Background()

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:17-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategyAndDeadline(180*time.Second,
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(1),
			wait.ForListeningPort("5432/tcp"),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	t.Cleanup(func() {
		container.Terminate(ctx)
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := runMigrations(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return &TestDB{DB: db, Container: container}
}

func runMigrations(db *sqlx.DB) error {
	goose.SetBaseFS(migrationsFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	if err := goose.Up(db.DB, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}

func TruncateTables(t *testing.T, db *sqlx.DB) {
	t.Helper()

	tables := []string{
		"collection_schedules",
		"prices",
		"hotel_reviews",
		"hotel_ratings",
		"hotels",
		"user_settings",
		"users",
		"vacations",
	}

	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			t.Fatalf("failed to truncate table %s: %v", table, err)
		}
	}
}

func TruncateExcept(t *testing.T, db *sqlx.DB, except ...string) {
	t.Helper()

	skip := make(map[string]bool, len(except))
	for _, s := range except {
		skip[s] = true
	}

	tables := []string{
		"collection_schedules",
		"prices",
		"hotel_reviews",
		"hotel_ratings",
		"hotels",
		"user_settings",
		"users",
		"vacations",
	}

	for _, table := range tables {
		if skip[table] {
			continue
		}
		if _, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			t.Fatalf("failed to truncate table %s: %v", table, err)
		}
	}
}

func init() {
	goose.SetLogger(goose.NopLogger())
}
