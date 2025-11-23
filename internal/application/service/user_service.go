package service

import (
	"context"
	"fmt"

	"reviewer-service/internal/domain/errors"
	"reviewer-service/internal/domain/interfaces"
	"reviewer-service/internal/domain/models"
	"reviewer-service/pkg/logger"
)

type UserService struct {
	userRepo interfaces.UserRepository
}

func NewUserService(userRepo interfaces.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

func (s *UserService) GetByID(ctx context.Context, userID string) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение пользователя", "user_id", userID)

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error("ошибка получения пользователя", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		log.Info("пользователь не найден", "user_id", userID)
		return nil, errors.ErrUserNotFound
	}

	return user, nil
}

func (s *UserService) SetActive(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	log := logger.FromContext(ctx)
	log.Debug("изменение статуса пользователя", "user_id", userID, "is_active", isActive)

	//проверка существования пользователя
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error("ошибка получения пользователя", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		log.Info("пользователь не найден", "user_id", userID)
		return nil, errors.ErrUserNotFound
	}

	//обновление статуста
	err = s.userRepo.SetActive(ctx, userID, isActive)
	if err != nil {
		log.Error("ошибка обновления статуса", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to set active status: %w", err)
	}

	//получение обновленного пользователя
	user, err = s.userRepo.GetByID(ctx, userID)
	if err != nil {
		log.Error("ошибка получения обновленного пользователя", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get updated user: %w", err)
	}

	log.Info("статус пользователя обновлен", "user_id", userID, "is_active", isActive)
	return user, nil
}

func (s *UserService) GetActiveByTeam(ctx context.Context, teamName string) ([]*models.User, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение активных пользователей команды", "team_name", teamName)

	users, err := s.userRepo.GetActiveByTeam(ctx, teamName)
	if err != nil {
		log.Error("ошибка получения активных пользователей", "error", err, "team_name", teamName)
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}

	return users, nil
}

func (s *UserService) DeactivateTeamAndReassign(ctx context.Context, teamName string, prService *PRService) error {
	log := logger.FromContext(ctx)
	log.Debug("массовая деактивация команды и переназначение PR", "team_name", teamName)

	users, err := s.userRepo.GetByTeam(ctx, teamName)
	if err != nil {
		log.Error("ошибка получения пользователей команды", "error", err, "team_name", teamName)
		return fmt.Errorf("failed to get team users: %w", err)
	}

	if len(users) == 0 {
		log.Info("пользователи команды не найдены", "team_name", teamName)
		return nil
	}

	deactivateIDs := make([]string, 0)
	for _, u := range users {
		if u.IsActive {
			deactivateIDs = append(deactivateIDs, u.ID)
		}
	}

	if len(deactivateIDs) == 0 {
		log.Info("нет активных пользователей для деактивации", "team_name", teamName)
		return nil
	}

	for _, id := range deactivateIDs {
		if err := s.userRepo.SetActive(ctx, id, false); err != nil {
			log.Error("ошибка деактивации пользователя", "error", err, "user_id", id)
			return fmt.Errorf("failed to deactivate user %s: %w", id, err)
		}
	}

	log.Info("пользователи деактивированы", "team_name", teamName, "count", len(deactivateIDs))

	for _, id := range deactivateIDs {
		if err := prService.ReassignForDeactivatedReviewer(ctx, id); err != nil {
			// не валим весь процесс из-за одного пользователя, просто логируем
			log.Error("ошибка переназначения PR для деактивированного ревьювера",
				"error", err,
				"user_id", id,
			)
		}
	}

	return nil
}
