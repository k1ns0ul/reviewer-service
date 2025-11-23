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

func TestUserService_SetActive(t *testing.T) {
	testDB := SetupTestDatabase(t)
	defer cleanupDatabase(t, testDB.DB)

	userRepo := repository.NewUserRepository(testDB.DB)
	teamRepo := repository.NewTeamRepository(testDB.DB)

	teamService := service.NewTeamService(teamRepo, userRepo)
	userService := service.NewUserService(userRepo)

	ctx := context.Background()

	members := []*models.TeamMember{
		{UserID: "user1", Username: "Alice", IsActive: true},
		{UserID: "user2", Username: "Bob", IsActive: false},
	}
	_, _, err := teamService.CreateTeam(ctx, "backend", members)
	require.NoError(t, err)

	t.Run("успешная деактивация пользователя", func(t *testing.T) {
		user, err := userService.SetActive(ctx, "user1", false)

		require.NoError(t, err)
		assert.False(t, user.IsActive)
	})

	t.Run("успешная активация пользователя", func(t *testing.T) {
		user, err := userService.SetActive(ctx, "user2", true)

		require.NoError(t, err)
		assert.True(t, user.IsActive)
	})

	t.Run("ошибка для несуществующего пользователя", func(t *testing.T) {
		_, err := userService.SetActive(ctx, "nonexistent", true)
		assert.Error(t, err)
	})
}

func TestUserService_GetActiveByTeam(t *testing.T) {
	testDB := SetupTestDatabase(t)
	defer cleanupDatabase(t, testDB.DB)

	userRepo := repository.NewUserRepository(testDB.DB)
	teamRepo := repository.NewTeamRepository(testDB.DB)

	teamService := service.NewTeamService(teamRepo, userRepo)
	userService := service.NewUserService(userRepo)

	ctx := context.Background()

	members := []*models.TeamMember{
		{UserID: "user1", Username: "Alice", IsActive: true},
		{UserID: "user2", Username: "Bob", IsActive: false},
		{UserID: "user3", Username: "Charlie", IsActive: true},
	}
	_, _, err := teamService.CreateTeam(ctx, "backend", members)
	require.NoError(t, err)

	t.Run("получение только активных пользователей", func(t *testing.T) {
		users, err := userService.GetActiveByTeam(ctx, "backend")

		require.NoError(t, err)
		assert.Len(t, users, 2)

		for _, user := range users {
			assert.True(t, user.IsActive)
		}
	})

	t.Run("пустой список для несуществующей команды", func(t *testing.T) {
		users, err := userService.GetActiveByTeam(ctx, "nonexistent")

		require.NoError(t, err)
		assert.Empty(t, users)
	})
}
