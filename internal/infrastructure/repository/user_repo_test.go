package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"reviewer-service/internal/domain/models"
	appLogger "reviewer-service/pkg/logger"
)

func setupUserRepoTest(t *testing.T) (*UserRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := &UserRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	cleanup := func() { db.Close() }
	return repo, mock, cleanup
}

func userTestContext(t *testing.T) context.Context {
	ctx := context.Background()
	log := appLogger.New(appLogger.Config{
		Level:  appLogger.LevelDebug,
		Format: "text",
		Output: nil,
	})
	return appLogger.WithLogger(ctx, log)
}

func TestUserRepository_Create(t *testing.T) {
	repo, mock, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := userTestContext(t)
	now := time.Now()
	user := &models.User{
		ID:        "user1",
		Username:  "Alice",
		TeamName:  "backend",
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.ID, user.Username, user.TeamName, user.IsActive, user.CreatedAt, user.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(ctx, user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_Found(t *testing.T) {
	repo, mock, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := userTestContext(t)
	now := time.Now()

	mock.ExpectQuery("SELECT (.+) FROM users").
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "username", "team_name", "is_active", "created_at", "updated_at",
		}).AddRow("user1", "Alice", "backend", true, now, now))

	user, err := repo.GetByID(ctx, "user1")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user1", user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := userTestContext(t)

	mock.ExpectQuery("SELECT (.+) FROM users").
		WithArgs("user404").
		WillReturnError(sql.ErrNoRows)

	user, err := repo.GetByID(ctx, "user404")
	assert.NoError(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetActiveByTeam(t *testing.T) {
	repo, mock, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := userTestContext(t)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "username", "team_name", "is_active", "created_at", "updated_at",
	}).
		AddRow("user1", "Alice", "backend", true, now, now).
		AddRow("user3", "Charlie", "backend", true, now, now)

	mock.ExpectQuery("SELECT (.+) FROM users").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows)

	users, err := repo.GetActiveByTeam(ctx, "backend")
	assert.NoError(t, err)
	assert.Len(t, users, 2)
	for _, u := range users {
		assert.True(t, u.IsActive)
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_SetActive(t *testing.T) {
	repo, mock, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := userTestContext(t)

	mock.ExpectExec("UPDATE users").
		WithArgs(true, "user1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.SetActive(ctx, "user1", true)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_Exists(t *testing.T) {
	repo, mock, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := userTestContext(t)

	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM users").
		WithArgs("user1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.Exists(ctx, "user1")
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByTeam_Error(t *testing.T) {
	repo, mock, cleanup := setupUserRepoTest(t)
	defer cleanup()

	ctx := userTestContext(t)

	mock.ExpectQuery("SELECT (.+) FROM users").
		WithArgs("backend").
		WillReturnError(errors.New("db error"))

	users, err := repo.GetByTeam(ctx, "backend")
	assert.Error(t, err)
	assert.Nil(t, users)
	assert.NoError(t, mock.ExpectationsWereMet())
}
