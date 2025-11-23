package interfaces

import (
	"context"
	"reviewer-service/internal/domain/models"
)

type TeamRepository interface {
	//создание новой команды
	Create(ctx context.Context, team *models.Team) error

	//получение команды по имени
	GetByName(ctx context.Context, name string) (*models.Team, error)

	//прповерка команды на существование
	Exists(ctx context.Context, name string) (bool, error)

	//получение мемберов команды
	GetMembers(ctx context.Context, teamName string) ([]*models.TeamMember, error)
}
