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

	"reviewer-service/internal/domain/enums"
	"reviewer-service/internal/domain/models"
	appLogger "reviewer-service/pkg/logger"
)

func setupPRRepoTest(t *testing.T) (*PRRepository, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := &PRRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}

	cleanup := func() { db.Close() }
	return repo, mock, cleanup
}

func prTestContext(t *testing.T) context.Context {
	ctx := context.Background()
	log := appLogger.New(appLogger.Config{
		Level:  appLogger.LevelDebug,
		Format: "text",
		Output: nil,
	})
	return appLogger.WithLogger(ctx, log)
}

func TestPRRepository_Create(t *testing.T) {
	repo, mock, cleanup := setupPRRepoTest(t)
	defer cleanup()

	ctx := prTestContext(t)
	now := time.Now()
	pr := &models.PullRequest{
		ID:        "pr-1",
		Name:      "Add feature",
		AuthorID:  "user1",
		Status:    enums.StatusOpen,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mock.ExpectExec("INSERT INTO pull_requests").
		WithArgs(pr.ID, pr.Name, pr.AuthorID, pr.Status, pr.CreatedAt, pr.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(ctx, pr)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPRRepository_Update(t *testing.T) {
	repo, mock, cleanup := setupPRRepoTest(t)
	defer cleanup()

	ctx := prTestContext(t)
	now := time.Now()
	pr := &models.PullRequest{
		ID:        "pr-1",
		Name:      "New name",
		AuthorID:  "user1",
		Status:    enums.StatusOpen,
		UpdatedAt: now,
		MergedAt:  nil,
	}

	mock.ExpectExec("UPDATE pull_requests").
		WithArgs(pr.Name, pr.Status, pr.UpdatedAt, pr.MergedAt, pr.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(ctx, pr)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPRRepository_GetByID_Found(t *testing.T) {
	repo, mock, cleanup := setupPRRepoTest(t)
	defer cleanup()

	ctx := prTestContext(t)
	now := time.Now()

	prRows := sqlmock.NewRows([]string{
		"id", "name", "author_id", "status", "created_at", "updated_at", "merged_at",
	}).AddRow("pr-1", "Add feature", "user1", enums.StatusOpen, now, now, nil)

	mock.ExpectQuery("SELECT (.+) FROM pull_requests").
		WithArgs("pr-1").
		WillReturnRows(prRows)

	reviewersRows := sqlmock.NewRows([]string{"reviewer_id"}).
		AddRow("user2").
		AddRow("user3")

	mock.ExpectQuery("SELECT (.+) FROM pr_reviewers").
		WithArgs("pr-1").
		WillReturnRows(reviewersRows)

	pr, err := repo.GetByID(ctx, "pr-1")
	assert.NoError(t, err)
	assert.NotNil(t, pr)
	assert.Equal(t, "pr-1", pr.ID)
	assert.Len(t, pr.AssignedReviewers, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPRRepository_GetByID_NotFound(t *testing.T) {
	repo, mock, cleanup := setupPRRepoTest(t)
	defer cleanup()

	ctx := prTestContext(t)

	mock.ExpectQuery("SELECT (.+) FROM pull_requests").
		WithArgs("pr-404").
		WillReturnError(sql.ErrNoRows)

	pr, err := repo.GetByID(ctx, "pr-404")
	assert.NoError(t, err)
	assert.Nil(t, pr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPRRepository_Exists(t *testing.T) {
	repo, mock, cleanup := setupPRRepoTest(t)
	defer cleanup()

	ctx := prTestContext(t)

	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM pull_requests").
		WithArgs("pr-1").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := repo.Exists(ctx, "pr-1")
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPRRepository_AssignReviewers(t *testing.T) {
	repo, mock, cleanup := setupPRRepoTest(t)
	defer cleanup()

	ctx := prTestContext(t)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO pr_reviewers").
		WithArgs("pr-1", "user2").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO pr_reviewers").
		WithArgs("pr-1", "user3").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.AssignReviewers(ctx, "pr-1", []string{"user2", "user3"})
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPRRepository_GetReviewers(t *testing.T) {
	repo, mock, cleanup := setupPRRepoTest(t)
	defer cleanup()

	ctx := prTestContext(t)

	rows := sqlmock.NewRows([]string{"reviewer_id"}).
		AddRow("user2").
		AddRow("user3")

	mock.ExpectQuery("SELECT (.+) FROM pr_reviewers").
		WithArgs("pr-1").
		WillReturnRows(rows)

	reviewers, err := repo.GetReviewers(ctx, "pr-1")
	assert.NoError(t, err)
	assert.Len(t, reviewers, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPRRepository_RemoveReviewer(t *testing.T) {
	repo, mock, cleanup := setupPRRepoTest(t)
	defer cleanup()

	ctx := prTestContext(t)

	mock.ExpectExec("DELETE FROM pr_reviewers").
		WithArgs("pr-1", "user2").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.RemoveReviewer(ctx, "pr-1", "user2")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPRRepository_GetByReviewer(t *testing.T) {
	repo, mock, cleanup := setupPRRepoTest(t)
	defer cleanup()

	ctx := prTestContext(t)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "name", "author_id", "status", "created_at", "updated_at", "merged_at",
	}).
		AddRow("pr-1", "Feature 1", "user1", enums.StatusOpen, now, now, nil).
		AddRow("pr-2", "Feature 2", "user1", enums.StatusOpen, now, now, nil)

	mock.ExpectQuery("SELECT DISTINCT pr.id").
		WithArgs("user2").
		WillReturnRows(rows)

	prs, err := repo.GetByReviewer(ctx, "user2")
	assert.NoError(t, err)
	assert.Len(t, prs, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPRRepository_Merge(t *testing.T) {
	repo, mock, cleanup := setupPRRepoTest(t)
	defer cleanup()

	ctx := prTestContext(t)

	mock.ExpectExec("UPDATE pull_requests").
		WithArgs("MERGED", "pr-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Merge(ctx, "pr-1")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
