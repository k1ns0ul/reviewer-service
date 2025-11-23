package repository

import (
	"context"
	"database/sql"
)

type StatsRepository struct {
	db *sql.DB
}

func NewStatsRepository(db *sql.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

type Stats struct {
	TotalPRs         int `json:"total_prs"`
	TotalMergedPRs   int `json:"total_merged_prs"`
	TotalAssignments int `json:"total_assignments"`
	TotalUsers       int `json:"total_users"`
	TotalActiveUsers int `json:"total_active_users"`
	TotalTeams       int `json:"total_teams"`
}

func (r *StatsRepository) GetStats(ctx context.Context) (*Stats, error) {
	s := &Stats{}

	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM pull_requests`).Scan(&s.TotalPRs); err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM pull_requests WHERE status = 'MERGED'`).Scan(&s.TotalMergedPRs); err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM pr_reviewers`).Scan(&s.TotalAssignments); err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&s.TotalUsers); err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE is_active = true`).Scan(&s.TotalActiveUsers); err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM teams`).Scan(&s.TotalTeams); err != nil {
		return nil, err
	}

	return s, nil
}
