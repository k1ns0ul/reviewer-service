package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"reviewer-service/internal/domain/constants"
	"reviewer-service/internal/domain/enums"
	"reviewer-service/internal/domain/errors"
	"reviewer-service/internal/domain/interfaces"
	"reviewer-service/internal/domain/models"
	"reviewer-service/pkg/logger"
)

type PRService struct {
	prRepo   interfaces.PRRepository
	userRepo interfaces.UserRepository
}

func NewPRService(prRepo interfaces.PRRepository, userRepo interfaces.UserRepository) *PRService {
	return &PRService{
		prRepo:   prRepo,
		userRepo: userRepo,
	}
}

func (s *PRService) CreatePR(ctx context.Context, prID, prName, authorID string) (*models.PullRequest, error) {
	log := logger.FromContext(ctx)
	log.Debug("создание PR", "pr_id", prID, "pr_name", prName, "author_id", authorID)

	exists, err := s.prRepo.Exists(ctx, prID)
	if err != nil {
		log.Error("ошибка проверки существования PR", "error", err, "pr_id", prID)
		return nil, fmt.Errorf("failed to check PR existence: %w", err)
	}

	if exists {
		log.Info("PR уже существует", "pr_id", prID)
		return nil, errors.ErrPRExists
	}

	author, err := s.userRepo.GetByID(ctx, authorID)
	if err != nil {
		log.Error("ошибка получения автора", "error", err, "author_id", authorID)
		return nil, fmt.Errorf("failed to get author: %w", err)
	}

	if author == nil {
		log.Info("автор не найден", "author_id", authorID)
		return nil, errors.ErrUserNotFound
	}

	pr := &models.PullRequest{
		ID:                prID,
		Name:              prName,
		AuthorID:          authorID,
		Status:            enums.StatusOpen,
		AssignedReviewers: make([]string, 0),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err = s.prRepo.Create(ctx, pr)
	if err != nil {
		log.Error("ошибка создания PR", "error", err, "pr_id", prID)
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	reviewers, err := s.assignInitialReviewers(ctx, authorID, author.TeamName)
	if err != nil {
		log.Error("ошибка назначения ревьюверов", "error", err, "pr_id", prID)
		return nil, fmt.Errorf("failed to assign reviewers: %w", err)
	}

	if len(reviewers) > 0 {
		err = s.prRepo.AssignReviewers(ctx, prID, reviewers)
		if err != nil {
			log.Error("ошибка сохранения ревьюверов", "error", err, "pr_id", prID)
			return nil, fmt.Errorf("failed to save reviewers: %w", err)
		}
		pr.AssignedReviewers = reviewers
	}

	log.Info("PR успешно создан", "pr_id", prID, "reviewers_count", len(reviewers))
	return pr, nil
}

func (s *PRService) MergePR(ctx context.Context, prID string) (*models.PullRequest, error) {
	log := logger.FromContext(ctx)
	log.Debug("мердж PR", "pr_id", prID)

	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		log.Error("ошибка получения PR", "error", err, "pr_id", prID)
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	if pr == nil {
		log.Info("PR не найден", "pr_id", prID)
		return nil, errors.ErrPRNotFound
	}

	//если уже смержен то возвращаем текущее состояние
	if pr.IsMerged() {
		log.Info("PR уже смержен", "pr_id", prID)
		return pr, nil
	}

	//мердж пр
	err = s.prRepo.Merge(ctx, prID)
	if err != nil {
		log.Error("ошибка мерджа PR", "error", err, "pr_id", prID)
		return nil, fmt.Errorf("failed to merge PR: %w", err)
	}

	pr, err = s.prRepo.GetByID(ctx, prID)
	if err != nil {
		log.Error("ошибка получения обновленного PR", "error", err, "pr_id", prID)
		return nil, fmt.Errorf("failed to get updated PR: %w", err)
	}

	log.Info("PR успешно смержен", "pr_id", prID)
	return pr, nil
}

func (s *PRService) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (*models.PullRequest, string, error) {
	log := logger.FromContext(ctx)
	log.Debug("переназначение ревьювера", "pr_id", prID, "old_reviewer_id", oldReviewerID)

	pr, err := s.prRepo.GetByID(ctx, prID)
	if err != nil {
		log.Error("ошибка получения PR", "error", err, "pr_id", prID)
		return nil, "", fmt.Errorf("failed to get PR: %w", err)
	}

	if pr == nil {
		log.Info("PR не найден", "pr_id", prID)
		return nil, "", errors.ErrPRNotFound
	}

	if pr.IsMerged() {
		log.Info("нельзя переназначить ревьювера на смерженном PR", "pr_id", prID)
		return nil, "", errors.ErrPRMerged
	}

	if !pr.IsReviewerAssigned(oldReviewerID) {
		log.Info("пользователь не назначен ревьювером", "pr_id", prID, "user_id", oldReviewerID)
		return nil, "", errors.ErrNotAssigned
	}

	oldReviewer, err := s.userRepo.GetByID(ctx, oldReviewerID)
	if err != nil {
		log.Error("ошибка получения ревьювера", "error", err, "reviewer_id", oldReviewerID)
		return nil, "", fmt.Errorf("failed to get reviewer: %w", err)
	}

	if oldReviewer == nil {
		log.Info("ревьювер не найден", "reviewer_id", oldReviewerID)
		return nil, "", errors.ErrUserNotFound
	}

	candidates, err := s.userRepo.GetActiveByTeam(ctx, oldReviewer.TeamName)
	if err != nil {
		log.Error("ошибка получения кандидатов", "error", err, "team_name", oldReviewer.TeamName)
		return nil, "", fmt.Errorf("failed to get candidates: %w", err)
	}

	//искление автора pr, старого ревьювера и текущих ревьюверов
	var availableCandidates []string
	for _, candidate := range candidates {
		if candidate.ID == pr.AuthorID || candidate.ID == oldReviewerID {
			continue
		}
		if pr.IsReviewerAssigned(candidate.ID) {
			continue
		}
		availableCandidates = append(availableCandidates, candidate.ID)
	}

	if len(availableCandidates) == 0 {
		log.Info("нет доступных кандидатов для замены", "pr_id", prID, "team_name", oldReviewer.TeamName)
		return nil, "", errors.ErrNoCandidate
	}

	//выбор случайного кандидата
	newReviewerID := availableCandidates[rand.Intn(len(availableCandidates))]

	//удаление старого ревьювера
	err = s.prRepo.RemoveReviewer(ctx, prID, oldReviewerID)
	if err != nil {
		log.Error("ошибка удаления ревьювера", "error", err, "pr_id", prID, "reviewer_id", oldReviewerID)
		return nil, "", fmt.Errorf("failed to remove reviewer: %w", err)
	}

	//назначение нового ревьювера
	err = s.prRepo.AssignReviewers(ctx, prID, []string{newReviewerID})
	if err != nil {
		log.Error("ошибка назначения нового ревьювера", "error", err, "pr_id", prID, "new_reviewer_id", newReviewerID)
		return nil, "", fmt.Errorf("failed to assign new reviewer: %w", err)
	}

	pr, err = s.prRepo.GetByID(ctx, prID)
	if err != nil {
		log.Error("ошибка получения обновленного PR", "error", err, "pr_id", prID)
		return nil, "", fmt.Errorf("failed to get updated PR: %w", err)
	}

	log.Info("ревьювер успешно переназначен",
		"pr_id", prID,
		"old_reviewer_id", oldReviewerID,
		"new_reviewer_id", newReviewerID,
	)

	return pr, newReviewerID, nil
}

func (s *PRService) GetPRsByReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequest, error) {
	log := logger.FromContext(ctx)
	log.Debug("получение PR по ревьюверу", "reviewer_id", reviewerID)

	exists, err := s.userRepo.Exists(ctx, reviewerID)
	if err != nil {
		log.Error("ошибка проверки существования пользователя", "error", err, "reviewer_id", reviewerID)
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	if !exists {
		log.Info("пользователь не найден", "reviewer_id", reviewerID)
		return nil, errors.ErrUserNotFound
	}

	//получение pr
	prs, err := s.prRepo.GetByReviewer(ctx, reviewerID)
	if err != nil {
		log.Error("ошибка получения PR", "error", err, "reviewer_id", reviewerID)
		return nil, fmt.Errorf("failed to get PRs: %w", err)
	}

	log.Debug("PR по ревьюверу получены", "reviewer_id", reviewerID, "count", len(prs))
	return prs, nil
}

// назначение до 2 случайных активных ревьюверов из команды автора
func (s *PRService) assignInitialReviewers(ctx context.Context, authorID, teamName string) ([]string, error) {
	log := logger.FromContext(ctx)

	candidates, err := s.userRepo.GetActiveByTeam(ctx, teamName)
	if err != nil {
		return nil, err
	}

	//фильтр
	var availableCandidates []string
	for _, candidate := range candidates {
		if candidate.ID != authorID {
			availableCandidates = append(availableCandidates, candidate.ID)
		}
	}

	if len(availableCandidates) == 0 {
		log.Info("нет доступных ревьюверов", "team_name", teamName)
		return []string{}, nil
	}

	//премешивание кандидатов
	rand.Shuffle(len(availableCandidates), func(i, j int) {
		availableCandidates[i], availableCandidates[j] = availableCandidates[j], availableCandidates[i]
	})

	//выбираем до max reviewers
	count := constants.MaxReviewers
	if len(availableCandidates) < count {
		count = len(availableCandidates)
	}

	reviewers := availableCandidates[:count]
	log.Debug("ревьюверы выбраны", "count", len(reviewers))

	return reviewers, nil
}

func (s *PRService) ReassignForDeactivatedReviewer(ctx context.Context, userID string) error {
	log := logger.FromContext(ctx)
	log.Debug("переназначение PR для деактивированного ревьювера", "user_id", userID)

	prs, err := s.prRepo.GetByReviewer(ctx, userID)
	if err != nil {
		log.Error("ошибка получения PR по ревьюверу", "error", err, "user_id", userID)
		return fmt.Errorf("failed to get PRs by reviewer: %w", err)
	}

	if len(prs) == 0 {
		log.Debug("PR для ревьювера не найдены", "user_id", userID)
		return nil
	}

	for _, pr := range prs {
		if pr.IsMerged() {
			continue
		}

		if !pr.IsReviewerAssigned(userID) {
			continue
		}

		_, _, err := s.ReassignReviewer(ctx, pr.ID, userID)
		if err != nil {
			// в рамках безопасного переназначения не валим всё, выводим только лог
			log.Error("ошибка переназначения ревьювера для PR при деактивации",
				"error", err,
				"pr_id", pr.ID,
				"old_reviewer_id", userID,
			)
			continue
		}
	}

	return nil
}
