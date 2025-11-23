package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"reviewer-service/internal/domain/models"
	appLogger "reviewer-service/pkg/logger"
)

func setupTeamRepoTest(t *testing.T) (*TeamRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := &TeamRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	cleanup := func() { db.Close() }
	return repo, mock, cleanup
}

func teamTestContext(t *testing.T) context.Context {
	ctx := context.Background()
	log := appLogger.New(appLogger.Config{
		Level:  appLogger.LevelDebug,
		Format: "text",
		Output: nil,
	})
	return appLogger.WithLogger(ctx, log)
}

func TestTeamRepository_Create(t *testing.T) {
	repo, mock, cleanup := setupTeamRepoTest(t)
	defer cleanup()

	ctx := teamTestContext(t)
	now := time.Now()
	team := &models.Team{
		Name:      "backend",
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO teams").
		WithArgs(team.Name, team.CreatedAt, team.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(ctx, team)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTeamRepository_GetByName_Found(t *testing.T) {
	repo, mock, cleanup := setupTeamRepoTest(t)
	defer cleanup()

	ctx := teamTestContext(t)
	now := time.Now()

	mock.ExpectQuery("SELECT (.+) FROM teams").
		WithArgs("backend").
		WillReturnRows(sqlmock.NewRows([]string{"name", "created_at", "updated_at"}).
			AddRow("backend", now, now))

	team, err := repo.GetByName(ctx, "backend")
	assert.NoError(t, err)
	assert.NotNil(t, team)
	assert.Equal(t, "backend", team.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTeamRepository_GetByName_NotFound(t *testing.T) {
	repo, mock, cleanup := setupTeamRepoTest(t)
	defer cleanup()

	ctx := teamTestContext(t)

	mock.ExpectQuery("SELECT (.+) FROM teams").
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	team, err := repo.GetByName(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, team)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTeamRepository_Exists(t *testing.T) {
	repo, mock, cleanup := setupTeamRepoTest(t)
	defer cleanup()

	ctx := teamTestContext(t)

	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM teams").
		WithArgs("backend").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.Exists(ctx, "backend")
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTeamRepository_GetMembers(t *testing.T) {
	repo, mock, cleanup := setupTeamRepoTest(t)
	defer cleanup()

	ctx := teamTestContext(t)

	rows := sqlmock.NewRows([]string{"id", "username", "is_active"}).
		AddRow("user1", "Alice", true).
		AddRow("user2", "Bob", false)

	mock.ExpectQuery("SELECT (.+) FROM users").
		WithArgs("backend").
		WillReturnRows(rows)

	members, err := repo.GetMembers(ctx, "backend")
	assert.NoError(t, err)
	assert.Len(t, members, 2)
	assert.Equal(t, "user1", members[0].UserID)
	assert.NoError(t, mock.ExpectationsWereMet())
}
