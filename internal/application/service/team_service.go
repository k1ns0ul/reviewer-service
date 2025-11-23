package service

import (
	"context"
	"fmt"
	"time"

	"reviewer-service/internal/domain/errors"
	"reviewer-service/internal/domain/interfaces"
	"reviewer-service/internal/domain/models"
	"reviewer-service/pkg/logger"
)

type TeamService struct {
	teamRepo interfaces.TeamRepository
	userRepo interfaces.UserRepository
}

func NewTeamService(teamRepo interfaces.TeamRepository, userRepo interfaces.UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, teamName string, members []*models.TeamMember) (*models.Team, []*models.TeamMember, error) {
	log := logger.FromContext(ctx)
	log.Debug("создание команды", "team_name", teamName, "members_count", len(members))

	exists, err := s.teamRepo.Exists(ctx, teamName)
	if err != nil {
		log.Error("ошибка проверки существования команды", "error", err, "team_name", teamName)
		return nil, nil, fmt.Errorf("failed to check team existence: %w", err)
	}

	if exists {
		log.Info("команда уже существует", "team_name", teamName)
		return nil, nil, errors.ErrTeamExists
	}

	//создание команды
	team := &models.Team{
		Name:      teamName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.teamRepo.Create(ctx, team)
	if err != nil {
		log.Error("ошибка создания команды", "error", err, "team_name", teamName)
		return nil, nil, fmt.Errorf("failed to create team: %w", err)
	}

	//создание / обнлвление пользователя
	now := time.Now()
	for _, member := range members {
		user := &models.User{
			ID:        member.UserID,
			Username:  member.Username,
			TeamName:  teamName,
			IsActive:  member.IsActive,
			CreatedAt: now,
			UpdatedAt: now,
		}

		err = s.userRepo.Create(ctx, user)
		if err != nil {
			log.Error("ошибка создания пользователя", "error", err, "user_id", member.UserID)
			return nil, nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	log.Info("команда успешно создана", "team_name", teamName, "members_count", len(members))
	return team, members, nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*models.Team, []*models.TeamMember, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение команды", "team_name", teamName)

	//получение команды
	team, err := s.teamRepo.GetByName(ctx, teamName)
	if err != nil {
		log.Error("ошибка получения команды", "error", err, "team_name", teamName)
		return nil, nil, fmt.Errorf("failed to get team: %w", err)
	}

	if team == nil {
		log.Info("команда не найдена", "team_name", teamName)
		return nil, nil, errors.ErrTeamNotFound
	}

	//получение участников
	members, err := s.teamRepo.GetMembers(ctx, teamName)
	if err != nil {
		log.Error("ошибка получения участников", "error", err, "team_name", teamName)
		return nil, nil, fmt.Errorf("failed to get team members: %w", err)
	}

	log.Debug("команда успешно получена", "team_name", teamName, "members_count", len(members))
	return team, members, nil
}
