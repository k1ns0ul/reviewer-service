package interfaces

import (
	"context"
	"reviewer-service/internal/domain/models"
)

type UserRepository interface {
	//создание новго пользователя
	Create(ctx context.Context, user *models.User) error

	//обновление дпнных пользователя
	Update(ctx context.Context, user *models.User) error

	//получение пользователя по айди
	GetByID(ctx context.Context, id string) (*models.User, error)

	//получение всех пользоателей команды
	GetByTeam(ctx context.Context, teamName string) ([]*models.User, error)

	//получение активных пользователей команды
	GetActiveByTeam(ctx context.Context, teamName string) ([]*models.User, error)

	//обновление статусп активности
	SetActive(ctx context.Context, userID string, isActive bool) error

	//проверка существования пользователя
	Exists(ctx context.Context, userID string) (bool, error)
}
