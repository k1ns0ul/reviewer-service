package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/lib/pq"
)

type TestDatabase struct {
	container *postgres.PostgresContainer
	DB        *sql.DB
	connStr   string
}

func SetupTestDatabase(t *testing.T) *TestDatabase {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("не удалось запустить контейнер: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("не удалось получить connection string: %v", err)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("не удалось подключиться к БД: %v", err)
	}

	if err := applyMigrations(db); err != nil {
		t.Fatalf("не удалось применить миграции: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("ошибка закрытия БД: %v", err)
		}
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Errorf("ошибка остановки контейнера: %v", err)
		}
	})

	return &TestDatabase{
		container: pgContainer,
		DB:        db,
		connStr:   connStr,
	}
}

func applyMigrations(db *sql.DB) error {
	migrationsPath := filepath.Join("..", "..", "migrations")

	migrations := []string{
		"000001_create_teams.up.sql",
		"000002_create_users.up.sql",
		"000003_create_pull_requests.up.sql",
		"000004_create_pr_reviewers.up.sql",
	}

	for _, migration := range migrations {
		migrationPath := filepath.Join(migrationsPath, migration)
		content, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("не удалось прочитать миграцию %s: %w", migration, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("не удалось применить миграцию %s: %w", migration, err)
		}
	}

	return nil
}

func cleanupDatabase(t *testing.T, db *sql.DB) {
	tables := []string{"pr_reviewers", "pull_requests", "users", "teams"}
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Logf("предупреждение: не удалось очистить таблицу %s: %v", table, err)
		}
	}
}
