package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatsRepository_GetStats(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewStatsRepository(db)
	ctx := context.Background()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM pull_requests`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(10))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM pull_requests WHERE status = 'MERGED'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(4))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM pr_reviewers`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(15))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(20))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM users WHERE is_active = true`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(12))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM teams`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	stats, err := repo.GetStats(ctx)
	require.NoError(t, err)

	assert.Equal(t, 10, stats.TotalPRs)
	assert.Equal(t, 4, stats.TotalMergedPRs)
	assert.Equal(t, 15, stats.TotalAssignments)
	assert.Equal(t, 20, stats.TotalUsers)
	assert.Equal(t, 12, stats.TotalActiveUsers)
	assert.Equal(t, 5, stats.TotalTeams)

	assert.NoError(t, mock.ExpectationsWereMet())
}
