package integration

import (
	"context"
	"testing"

	"reviewer-service/internal/application/service"
	"reviewer-service/internal/domain/models"
	"reviewer-service/internal/infrastructure/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamService_CreateTeam(t *testing.T) {
	testDB := SetupTestDatabase(t)
	defer cleanupDatabase(t, testDB.DB)

	teamRepo := repository.NewTeamRepository(testDB.DB)
	userRepo := repository.NewUserRepository(testDB.DB)
	teamService := service.NewTeamService(teamRepo, userRepo)

	ctx := context.Background()

	t.Run("успешное создание команды", func(t *testing.T) {
		members := []*models.TeamMember{
			{UserID: "user1", Username: "Alice", IsActive: true},
			{UserID: "user2", Username: "Bob", IsActive: true},
		}

		team, resultMembers, err := teamService.CreateTeam(ctx, "backend", members)

		require.NoError(t, err)
		assert.Equal(t, "backend", team.Name)
		assert.Len(t, resultMembers, 2)
	})

	t.Run("ошибка при создании существующей команды", func(t *testing.T) {
		members := []*models.TeamMember{
			{UserID: "user3", Username: "Charlie", IsActive: true},
		}

		_, _, err := teamService.CreateTeam(ctx, "frontend", members)
		require.NoError(t, err)

		_, _, err = teamService.CreateTeam(ctx, "frontend", members)
		assert.Error(t, err)
	})
}

func TestTeamService_GetTeam(t *testing.T) {
	testDB := SetupTestDatabase(t)
	defer cleanupDatabase(t, testDB.DB)

	teamRepo := repository.NewTeamRepository(testDB.DB)
	userRepo := repository.NewUserRepository(testDB.DB)
	teamService := service.NewTeamService(teamRepo, userRepo)

	ctx := context.Background()

	members := []*models.TeamMember{
		{UserID: "user1", Username: "Alice", IsActive: true},
		{UserID: "user2", Username: "Bob", IsActive: false},
	}
	_, _, err := teamService.CreateTeam(ctx, "backend", members)
	require.NoError(t, err)

	t.Run("успешное получение команды", func(t *testing.T) {
		team, resultMembers, err := teamService.GetTeam(ctx, "backend")

		require.NoError(t, err)
		assert.Equal(t, "backend", team.Name)
		assert.Len(t, resultMembers, 2)
	})

	t.Run("ошибка при получении несуществующей команды", func(t *testing.T) {
		_, _, err := teamService.GetTeam(ctx, "nonexistent")
		assert.Error(t, err)
	})
}
