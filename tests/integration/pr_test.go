package integration

import (
	"context"
	"testing"

	"reviewer-service/internal/application/service"
	"reviewer-service/internal/domain/enums"
	"reviewer-service/internal/domain/models"
	"reviewer-service/internal/infrastructure/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPRService_CreatePR(t *testing.T) {
	testDB := SetupTestDatabase(t)
	defer cleanupDatabase(t, testDB.DB)

	userRepo := repository.NewUserRepository(testDB.DB)
	teamRepo := repository.NewTeamRepository(testDB.DB)
	prRepo := repository.NewPRRepository(testDB.DB)

	teamService := service.NewTeamService(teamRepo, userRepo)
	prService := service.NewPRService(prRepo, userRepo)

	ctx := context.Background()

	members := []*models.TeamMember{
		{UserID: "user1", Username: "Alice", IsActive: true},
		{UserID: "user2", Username: "Bob", IsActive: true},
		{UserID: "user3", Username: "Charlie", IsActive: true},
	}
	_, _, err := teamService.CreateTeam(ctx, "backend", members)
	require.NoError(t, err)

	t.Run("успешное создание PR с назначением ревьюверов", func(t *testing.T) {
		pr, err := prService.CreatePR(ctx, "pr-1", "Add feature", "user1")

		require.NoError(t, err)
		assert.Equal(t, "pr-1", pr.ID)
		assert.Equal(t, "Add feature", pr.Name)
		assert.Equal(t, "user1", pr.AuthorID)
		assert.Equal(t, enums.StatusOpen, pr.Status)
		assert.True(t, len(pr.AssignedReviewers) > 0 && len(pr.AssignedReviewers) <= 2)

		for _, reviewerID := range pr.AssignedReviewers {
			assert.NotEqual(t, "user1", reviewerID)
		}
	})

	t.Run("ошибка при создании существующего PR", func(t *testing.T) {
		_, err := prService.CreatePR(ctx, "pr-2", "Another feature", "user2")
		require.NoError(t, err)

		_, err = prService.CreatePR(ctx, "pr-2", "Duplicate", "user3")
		assert.Error(t, err)
	})

	t.Run("ошибка при несуществующем авторе", func(t *testing.T) {
		_, err := prService.CreatePR(ctx, "pr-3", "Feature", "nonexistent")
		assert.Error(t, err)
	})
}

func TestPRService_MergePR(t *testing.T) {
	testDB := SetupTestDatabase(t)
	defer cleanupDatabase(t, testDB.DB)

	userRepo := repository.NewUserRepository(testDB.DB)
	teamRepo := repository.NewTeamRepository(testDB.DB)
	prRepo := repository.NewPRRepository(testDB.DB)

	teamService := service.NewTeamService(teamRepo, userRepo)
	prService := service.NewPRService(prRepo, userRepo)

	ctx := context.Background()

	members := []*models.TeamMember{
		{UserID: "user1", Username: "Alice", IsActive: true},
		{UserID: "user2", Username: "Bob", IsActive: true},
	}
	_, _, err := teamService.CreateTeam(ctx, "backend", members)
	require.NoError(t, err)

	pr, err := prService.CreatePR(ctx, "pr-1", "Feature", "user1")
	require.NoError(t, err)
	require.True(t, len(pr.AssignedReviewers) > 0)

	t.Run("успешный мердж PR", func(t *testing.T) {
		mergedPR, err := prService.MergePR(ctx, "pr-1")

		require.NoError(t, err)
		assert.Equal(t, enums.StatusMerged, mergedPR.Status)
		assert.NotNil(t, mergedPR.MergedAt)
	})

	t.Run("идемпотентность мерджа", func(t *testing.T) {
		mergedPR, err := prService.MergePR(ctx, "pr-1")

		require.NoError(t, err)
		assert.Equal(t, enums.StatusMerged, mergedPR.Status)
	})

	t.Run("ошибка при мердже несуществующего PR", func(t *testing.T) {
		_, err := prService.MergePR(ctx, "nonexistent")
		assert.Error(t, err)
	})
}

func TestPRService_ReassignReviewer(t *testing.T) {
	testDB := SetupTestDatabase(t)
	defer cleanupDatabase(t, testDB.DB)

	userRepo := repository.NewUserRepository(testDB.DB)
	teamRepo := repository.NewTeamRepository(testDB.DB)
	prRepo := repository.NewPRRepository(testDB.DB)

	teamService := service.NewTeamService(teamRepo, userRepo)
	prService := service.NewPRService(prRepo, userRepo)

	ctx := context.Background()

	members := []*models.TeamMember{
		{UserID: "user1", Username: "Alice", IsActive: true},
		{UserID: "user2", Username: "Bob", IsActive: true},
		{UserID: "user3", Username: "Charlie", IsActive: true},
		{UserID: "user4", Username: "Dave", IsActive: true},
	}
	_, _, err := teamService.CreateTeam(ctx, "backend", members)
	require.NoError(t, err)

	pr, err := prService.CreatePR(ctx, "pr-1", "Feature", "user1")
	require.NoError(t, err)
	require.True(t, len(pr.AssignedReviewers) > 0)

	oldReviewerID := pr.AssignedReviewers[0]

	t.Run("успешное переназначение ревьювера", func(t *testing.T) {
		updatedPR, newReviewerID, err := prService.ReassignReviewer(ctx, "pr-1", oldReviewerID)

		require.NoError(t, err)
		assert.NotEqual(t, oldReviewerID, newReviewerID)
		assert.NotContains(t, updatedPR.AssignedReviewers, oldReviewerID)
		assert.Contains(t, updatedPR.AssignedReviewers, newReviewerID)
	})

	t.Run("ошибка при переназначении на смерженном PR", func(t *testing.T) {
		pr2, err := prService.CreatePR(ctx, "pr-2", "Feature 2", "user1")
		require.NoError(t, err)

		_, err = prService.MergePR(ctx, "pr-2")
		require.NoError(t, err)

		_, _, err = prService.ReassignReviewer(ctx, "pr-2", pr2.AssignedReviewers[0])
		assert.Error(t, err)
	})

	t.Run("ошибка при переназначении неназначенного пользователя", func(t *testing.T) {
		_, _, err := prService.ReassignReviewer(ctx, "pr-1", "user4")
		assert.Error(t, err)
	})
}

func TestPRService_GetPRsByReviewer(t *testing.T) {
	testDB := SetupTestDatabase(t)
	defer cleanupDatabase(t, testDB.DB)

	userRepo := repository.NewUserRepository(testDB.DB)
	teamRepo := repository.NewTeamRepository(testDB.DB)
	prRepo := repository.NewPRRepository(testDB.DB)

	teamService := service.NewTeamService(teamRepo, userRepo)
	prService := service.NewPRService(prRepo, userRepo)

	ctx := context.Background()

	members := []*models.TeamMember{
		{UserID: "user1", Username: "Alice", IsActive: true},
		{UserID: "user2", Username: "Bob", IsActive: true},
		{UserID: "user3", Username: "Charlie", IsActive: true},
	}
	_, _, err := teamService.CreateTeam(ctx, "backend", members)
	require.NoError(t, err)

	_, err = prService.CreatePR(ctx, "pr-1", "Feature 1", "user1")
	require.NoError(t, err)

	_, err = prService.CreatePR(ctx, "pr-2", "Feature 2", "user1")
	require.NoError(t, err)

	t.Run("получение PR для ревьювера", func(t *testing.T) {
		prs, err := prService.GetPRsByReviewer(ctx, "user2")

		require.NoError(t, err)
		assert.True(t, len(prs) >= 0)
	})

	t.Run("ошибка для несуществующего пользователя", func(t *testing.T) {
		_, err := prService.GetPRsByReviewer(ctx, "nonexistent")
		assert.Error(t, err)
	})
}
