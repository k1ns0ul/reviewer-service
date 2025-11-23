package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"reviewer-service/internal/domain/interfaces"
	"reviewer-service/internal/domain/models"
	"reviewer-service/pkg/logger"

	"github.com/Masterminds/squirrel"
)

type PRRepository struct {
	db *sql.DB
	sb squirrel.StatementBuilderType
}

func NewPRRepository(db *sql.DB) interfaces.PRRepository {
	return &PRRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *PRRepository) Create(ctx context.Context, pr *models.PullRequest) error {
	log := logger.FromContext(ctx)
	log.Debug("создание PR",
		"pr_id", pr.ID,
		"pr_name", pr.Name,
		"author_id", pr.AuthorID,
	)

	query := r.sb.Insert("pull_requests").
		Columns("id", "name", "author_id", "status", "created_at", "updated_at").
		Values(pr.ID, pr.Name, pr.AuthorID, pr.Status, pr.CreatedAt, pr.UpdatedAt)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return fmt.Errorf("error building SQL: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error("ошибка создания PR", "error", err, "pr_id", pr.ID)
		return fmt.Errorf("error creating pull request: %w", err)
	}

	log.Info("PR успешно создан", "pr_id", pr.ID)
	return nil
}

func (r *PRRepository) Update(ctx context.Context, pr *models.PullRequest) error {
	log := logger.FromContext(ctx)
	log.Debug("обновление PR", "pr_id", pr.ID)

	query := r.sb.Update("pull_requests").
		Set("name", pr.Name).
		Set("status", pr.Status).
		Set("updated_at", pr.UpdatedAt).
		Set("merged_at", pr.MergedAt).
		Where(squirrel.Eq{"id": pr.ID})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return fmt.Errorf("error building SQL: %w", err)
	}

	result, err := r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error("ошибка обновления PR", "error", err, "pr_id", pr.ID)
		return fmt.Errorf("error updating pull request: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Info("PR успешно обновлен", "pr_id", pr.ID, "rows_affected", rowsAffected)

	return nil
}

func (r *PRRepository) GetByID(ctx context.Context, id string) (*models.PullRequest, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение PR по ID", "pr_id", id)

	query := r.sb.Select("id", "name", "author_id", "status", "created_at", "updated_at", "merged_at").
		From("pull_requests").
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	pr := &models.PullRequest{}
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status,
		&pr.CreatedAt, &pr.UpdatedAt, &pr.MergedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info("PR не найден", "pr_id", id)
			return nil, nil
		}
		log.Error("ошибка получения PR", "error", err, "pr_id", id)
		return nil, fmt.Errorf("error getting pull request: %w", err)
	}

	// Получаем ревьюверов
	reviewers, err := r.GetReviewers(ctx, id)
	if err != nil {
		return nil, err
	}
	pr.AssignedReviewers = reviewers

	log.Debug("PR успешно получен", "pr_id", pr.ID)
	return pr, nil
}

func (r *PRRepository) Exists(ctx context.Context, id string) (bool, error) {
	log := logger.FromContext(ctx)
	log.Debug("проверка существования PR", "pr_id", id)

	sqlQuery := "SELECT EXISTS(SELECT 1 FROM pull_requests WHERE id = $1)"

	var exists bool
	err := r.db.QueryRowContext(ctx, sqlQuery, id).Scan(&exists)
	if err != nil {
		log.Error("ошибка проверки существования", "error", err, "pr_id", id)
		return false, fmt.Errorf("error checking PR existence: %w", err)
	}

	return exists, nil
}

func (r *PRRepository) AssignReviewers(ctx context.Context, prID string, reviewerIDs []string) error {
	log := logger.FromContext(ctx)
	log.Debug("назначение ревьюверов", "pr_id", prID, "reviewers", reviewerIDs)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Error("ошибка начала транзакции", "error", err)
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	for _, reviewerID := range reviewerIDs {
		query := r.sb.Insert("pr_reviewers").
			Columns("pull_request_id", "reviewer_id", "assigned_at").
			Values(prID, reviewerID, squirrel.Expr("NOW()")).
			Suffix("ON CONFLICT (pull_request_id, reviewer_id) DO NOTHING")

		sqlQuery, args, err := query.ToSql()
		if err != nil {
			log.Error("ошибка построения SQL", "error", err)
			return fmt.Errorf("error building SQL: %w", err)
		}

		_, err = tx.ExecContext(ctx, sqlQuery, args...)
		if err != nil {
			log.Error("ошибка назначения ревьювера", "error", err, "reviewer_id", reviewerID)
			return fmt.Errorf("error assigning reviewer: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Error("ошибка коммита транзакции", "error", err)
		return fmt.Errorf("error committing transaction: %w", err)
	}

	log.Info("ревьюверы успешно назначены", "pr_id", prID, "count", len(reviewerIDs))
	return nil
}

func (r *PRRepository) GetReviewers(ctx context.Context, prID string) ([]string, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение ревьюверов PR", "pr_id", prID)

	query := r.sb.Select("reviewer_id").
		From("pr_reviewers").
		Where(squirrel.Eq{"pull_request_id": prID}).
		OrderBy("assigned_at")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error("ошибка получения ревьюверов", "error", err, "pr_id", prID)
		return nil, fmt.Errorf("error getting reviewers: %w", err)
	}
	defer rows.Close()

	reviewers := make([]string, 0)
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			log.Error("ошибка сканирования строки", "error", err)
			return nil, fmt.Errorf("error scanning reviewer row: %w", err)
		}
		reviewers = append(reviewers, reviewerID)
	}

	log.Debug("ревьюверы получены", "pr_id", prID, "count", len(reviewers))
	return reviewers, rows.Err()
}

func (r *PRRepository) RemoveReviewer(ctx context.Context, prID, reviewerID string) error {
	log := logger.FromContext(ctx)
	log.Debug("удаление ревьювера", "pr_id", prID, "reviewer_id", reviewerID)

	query := r.sb.Delete("pr_reviewers").
		Where(squirrel.Eq{"pull_request_id": prID, "reviewer_id": reviewerID})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return fmt.Errorf("error building SQL: %w", err)
	}

	result, err := r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error("ошибка удаления ревьювера", "error", err, "pr_id", prID, "reviewer_id", reviewerID)
		return fmt.Errorf("error removing reviewer: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Info("ревьювер удален", "pr_id", prID, "reviewer_id", reviewerID, "rows_affected", rowsAffected)

	return nil
}

func (r *PRRepository) GetByReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequest, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение PR по ревьюверу", "reviewer_id", reviewerID)

	query := r.sb.Select("DISTINCT pr.id", "pr.name", "pr.author_id", "pr.status", "pr.created_at", "pr.updated_at", "pr.merged_at").
		From("pull_requests pr").
		Join("pr_reviewers prr ON pr.id = prr.pull_request_id").
		Where(squirrel.Eq{"prr.reviewer_id": reviewerID}).
		OrderBy("pr.created_at DESC")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error("ошибка получения PR", "error", err, "reviewer_id", reviewerID)
		return nil, fmt.Errorf("error getting PRs by reviewer: %w", err)
	}
	defer rows.Close()

	prs := make([]*models.PullRequest, 0)
	for rows.Next() {
		pr := &models.PullRequest{}
		if err := rows.Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status,
			&pr.CreatedAt, &pr.UpdatedAt, &pr.MergedAt); err != nil {
			log.Error("ошибка сканирования строки", "error", err)
			return nil, fmt.Errorf("error scanning PR row: %w", err)
		}
		prs = append(prs, pr)
	}

	log.Info("PR по ревьюверу получены", "reviewer_id", reviewerID, "count", len(prs))
	return prs, rows.Err()
}

func (r *PRRepository) Merge(ctx context.Context, prID string) error {
	log := logger.FromContext(ctx)
	log.Debug("мердж PR", "pr_id", prID)

	query := r.sb.Update("pull_requests").
		Set("status", "MERGED").
		Set("merged_at", squirrel.Expr("NOW()")).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": prID})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return fmt.Errorf("error building SQL: %w", err)
	}

	result, err := r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error("ошибка мерджа PR", "error", err, "pr_id", prID)
		return fmt.Errorf("error merging PR: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Info("PR успешно смержен", "pr_id", prID, "rows_affected", rowsAffected)

	return nil
}
