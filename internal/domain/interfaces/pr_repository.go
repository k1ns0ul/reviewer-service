package interfaces

import (
	"context"
	"reviewer-service/internal/domain/models"
)

type PRRepository interface {
	//создание новго PR
	Create(ctx context.Context, pr *models.PullRequest) error

	//обновление PR
	Update(ctx context.Context, pr *models.PullRequest) error

	//получение PR по ID
	GetByID(ctx context.Context, id string) (*models.PullRequest, error)

	//проверка существования PR
	Exists(ctx context.Context, id string) (bool, error)

	//назначение ревьюверов на PR
	AssignReviewers(ctx context.Context, prID string, reviewerIDs []string) error

	//получение списка ревьюверов PR
	GetReviewers(ctx context.Context, prID string) ([]string, error)

	//удаление ревьювера из PR
	RemoveReviewer(ctx context.Context, prID, reviewerID string) error

	//получение все Pr где пользователь ревьюер
	GetByReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequest, error)

	//пометка смерженного PRа
	Merge(ctx context.Context, prID string) error
}
