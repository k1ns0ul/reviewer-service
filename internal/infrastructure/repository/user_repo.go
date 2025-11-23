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

type UserRepository struct {
	db *sql.DB
	sb squirrel.StatementBuilderType
}

func NewUserRepository(db *sql.DB) interfaces.UserRepository {
	return &UserRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	log := logger.FromContext(ctx)
	log.Debug("создание пользователя",
		"user_id", user.ID,
		"username", user.Username,
		"team_name", user.TeamName,
	)

	query := r.sb.Insert("users").
		Columns("id", "username", "team_name", "is_active", "created_at", "updated_at").
		Values(user.ID, user.Username, user.TeamName, user.IsActive, user.CreatedAt, user.UpdatedAt).
		Suffix(`ON CONFLICT (id) DO UPDATE SET 
			username = EXCLUDED.username,
			team_name = EXCLUDED.team_name,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at`)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return fmt.Errorf("error building SQL: %w", err)
	}

	_, err = r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error("ошибка создания пользователя", "error", err, "user_id", user.ID)
		return fmt.Errorf("error creating user: %w", err)
	}

	log.Info("пользователь успешно создан", "user_id", user.ID)
	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	return r.Create(ctx, user)
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение пользователя по ID", "user_id", id)

	query := r.sb.Select("id", "username", "team_name", "is_active", "created_at", "updated_at").
		From("users").
		Where(squirrel.Eq{"id": id})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	user := &models.User{}
	err = r.db.QueryRowContext(ctx, sqlQuery, args...).Scan(
		&user.ID, &user.Username, &user.TeamName, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info("пользователь не найден", "user_id", id)
			return nil, nil
		}
		log.Error("ошибка получения пользователя", "error", err, "user_id", id)
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	log.Debug("пользователь успешно получен", "user_id", user.ID)
	return user, nil
}

func (r *UserRepository) GetByTeam(ctx context.Context, teamName string) ([]*models.User, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение пользователей команды", "team_name", teamName)

	query := r.sb.Select("id", "username", "team_name", "is_active", "created_at", "updated_at").
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
		log.Error("ошибка получения пользователей", "error", err, "team_name", teamName)
		return nil, fmt.Errorf("error getting users: %w", err)
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive,
			&user.CreatedAt, &user.UpdatedAt); err != nil {
			log.Error("ошибка сканирования строки", "error", err)
			return nil, fmt.Errorf("error scanning user row: %w", err)
		}
		users = append(users, user)
	}

	log.Info("пользователи команды получены", "team_name", teamName, "count", len(users))
	return users, rows.Err()
}

func (r *UserRepository) GetActiveByTeam(ctx context.Context, teamName string) ([]*models.User, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение активных пользователей команды", "team_name", teamName)

	query := r.sb.Select("id", "username", "team_name", "is_active", "created_at", "updated_at").
		From("users").
		Where(squirrel.Eq{"team_name": teamName, "is_active": true}).
		OrderBy("username")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error("ошибка получения активных пользователей", "error", err, "team_name", teamName)
		return nil, fmt.Errorf("error getting active users: %w", err)
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive,
			&user.CreatedAt, &user.UpdatedAt); err != nil {
			log.Error("ошибка сканирования строки", "error", err)
			return nil, fmt.Errorf("error scanning user row: %w", err)
		}
		users = append(users, user)
	}

	log.Info("активные пользователи команды получены", "team_name", teamName, "count", len(users))
	return users, rows.Err()
}

func (r *UserRepository) SetActive(ctx context.Context, userID string, isActive bool) error {
	log := logger.FromContext(ctx)
	log.Debug("обновление статуса пользователя", "user_id", userID, "is_active", isActive)

	query := r.sb.Update("users").
		Set("is_active", isActive).
		Set("updated_at", squirrel.Expr("NOW()")).
		Where(squirrel.Eq{"id": userID})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		log.Error("ошибка построения SQL", "error", err)
		return fmt.Errorf("error building SQL: %w", err)
	}

	result, err := r.db.ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		log.Error("ошибка обновления статуса", "error", err, "user_id", userID)
		return fmt.Errorf("error updating user active status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Info("статус пользователя обновлен", "user_id", userID, "is_active", isActive, "rows_affected", rowsAffected)

	return nil
}

func (r *UserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	log := logger.FromContext(ctx)
	log.Debug("проверка существования пользователя", "user_id", userID)

	sqlQuery := "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)"

	var exists bool
	err := r.db.QueryRowContext(ctx, sqlQuery, userID).Scan(&exists)
	if err != nil {
		log.Error("ошибка проверки существования", "error", err, "user_id", userID)
		return false, fmt.Errorf("error checking user existence: %w", err)
	}

	return exists, nil
}
