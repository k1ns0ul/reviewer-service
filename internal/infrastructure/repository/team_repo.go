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

type TeamRepository struct {
	db *sql.DB
	sb squirrel.StatementBuilderType
}

func NewTeamRepository(db *sql.DB) interfaces.TeamRepository {
	return &TeamRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *TeamRepository) Create(ctx context.Context, team *models.Team) error {
	log := logger.FromContext(ctx)
	log.Debug("создание команды", "team_name", team.Name)

	query := r.sb.Insert("teams").
		Columns("name", "created_at", "updated_at").
		Values(team.Name, team.CreatedAt, team.UpdatedAt)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return fmt.Errorf("error building SQL: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error("ошибка создания команды", "error", err, "team_name", team.Name)
		return fmt.Errorf("error creating team: %w", err)
	}

	log.Info("команда успешно создана", "team_name", team.Name)
	return nil
}

func (r *TeamRepository) GetByName(ctx context.Context, name string) (*models.Team, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение команды по имени", "team_name", name)

	query := r.sb.Select("name", "created_at", "updated_at").
		From("teams").
		Where(squirrel.Eq{"name": name})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	team := &models.Team{}
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&team.Name, &team.CreatedAt, &team.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info("команда не найдена", "team_name", name)
			return nil, nil
		}
		log.Error("ошибка получения команды", "error", err, "team_name", name)
		return nil, fmt.Errorf("error getting team: %w", err)
	}

	log.Debug("команда успешно получена", "team_name", team.Name)
	return team, nil
}

func (r *TeamRepository) Exists(ctx context.Context, name string) (bool, error) {
	log := logger.FromContext(ctx)
	log.Debug("проверка существования команды", "team_name", name)

	sqlQuery := "SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)"

	var exists bool
	err := r.db.QueryRowContext(ctx, sqlQuery, name).Scan(&exists)
	if err != nil {
		log.Error("ошибка проверки существования", "error", err, "team_name", name)
		return false, fmt.Errorf("error checking team existence: %w", err)
	}

	return exists, nil
}

func (r *TeamRepository) GetMembers(ctx context.Context, teamName string) ([]*models.TeamMember, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение участников команды", "team_name", teamName)

	query := r.sb.Select("id", "username", "is_active").
		From("users").
		Where(squirrel.Eq{"team_name": teamName}).
		OrderBy("username")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error("ошибка получения участников", "error", err, "team_name", teamName)
		return nil, fmt.Errorf("error getting team members: %w", err)
	}
	defer rows.Close()

	members := make([]*models.TeamMember, 0)
	for rows.Next() {
		member := &models.TeamMember{}
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			log.Error("ошибка сканирования строки", "error", err)
			return nil, fmt.Errorf("error scanning member row: %w", err)
		}
		members = append(members, member)
	}

	log.Info("участники команды получены", "team_name", teamName, "count", len(members))
	return members, rows.Err()
}
