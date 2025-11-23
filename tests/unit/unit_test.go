package unit

import (
	"context"
	"testing"
	"time"

	"reviewer-service/internal/application/service"
	"reviewer-service/internal/domain/enums"
	domainErrors "reviewer-service/internal/domain/errors"
	"reviewer-service/internal/domain/models"

	"net/http"
	"net/http/httptest"
	"reviewer-service/internal/entrypoints/http/handler"
	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type TeamRepositoryMock struct {
	mock.Mock
}

func (m *TeamRepositoryMock) Create(ctx context.Context, team *models.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *TeamRepositoryMock) GetByName(ctx context.Context, name string) (*models.Team, error) {
	args := m.Called(ctx, name)
	if t, ok := args.Get(0).(*models.Team); ok {
		return t, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *TeamRepositoryMock) Exists(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *TeamRepositoryMock) GetMembers(ctx context.Context, teamName string) ([]*models.TeamMember, error) {
	args := m.Called(ctx, teamName)
	if v, ok := args.Get(0).([]*models.TeamMember); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

type UserRepositoryMock struct {
	mock.Mock
}

func (m *UserRepositoryMock) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *UserRepositoryMock) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *UserRepositoryMock) GetByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	if u, ok := args.Get(0).(*models.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *UserRepositoryMock) GetByTeam(ctx context.Context, teamName string) ([]*models.User, error) {
	args := m.Called(ctx, teamName)
	if v, ok := args.Get(0).([]*models.User); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *UserRepositoryMock) GetActiveByTeam(ctx context.Context, teamName string) ([]*models.User, error) {
	args := m.Called(ctx, teamName)
	if v, ok := args.Get(0).([]*models.User); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *UserRepositoryMock) SetActive(ctx context.Context, userID string, isActive bool) error {
	args := m.Called(ctx, userID, isActive)
	return args.Error(0)
}

func (m *UserRepositoryMock) Exists(ctx context.Context, userID string) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

type PRRepositoryMock struct {
	mock.Mock
}

func (m *PRRepositoryMock) Create(ctx context.Context, pr *models.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *PRRepositoryMock) Update(ctx context.Context, pr *models.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *PRRepositoryMock) GetByID(ctx context.Context, id string) (*models.PullRequest, error) {
	args := m.Called(ctx, id)
	if pr, ok := args.Get(0).(*models.PullRequest); ok {
		return pr, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *PRRepositoryMock) Exists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *PRRepositoryMock) AssignReviewers(ctx context.Context, prID string, reviewerIDs []string) error {
	args := m.Called(ctx, prID, reviewerIDs)
	return args.Error(0)
}

func (m *PRRepositoryMock) GetReviewers(ctx context.Context, prID string) ([]string, error) {
	args := m.Called(ctx, prID)
	if v, ok := args.Get(0).([]string); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *PRRepositoryMock) RemoveReviewer(ctx context.Context, prID, reviewerID string) error {
	args := m.Called(ctx, prID, reviewerID)
	return args.Error(0)
}

func (m *PRRepositoryMock) GetByReviewer(ctx context.Context, reviewerID string) ([]*models.PullRequest, error) {
	args := m.Called(ctx, reviewerID)
	if v, ok := args.Get(0).([]*models.PullRequest); ok {
		return v, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *PRRepositoryMock) Merge(ctx context.Context, prID string) error {
	args := m.Called(ctx, prID)
	return args.Error(0)
}

func TestTeamService_CreateTeam(t *testing.T) {
	ctx := context.Background()

	teamRepo := new(TeamRepositoryMock)
	userRepo := new(UserRepositoryMock)
	s := service.NewTeamService(teamRepo, userRepo)

	members := []*models.TeamMember{
		{UserID: "user1", Username: "Alice", IsActive: true},
		{UserID: "user2", Username: "Bob", IsActive: true},
	}

	t.Run("успешное создание команды", func(t *testing.T) {
		teamRepo.
			On("Exists", ctx, "backend").
			Return(false, nil).
			Once()

		teamRepo.
			On("Create", ctx, mock.MatchedBy(func(team *models.Team) bool {
				return team.Name == "backend"
			})).
			Return(nil).
			Once()

		userRepo.
			On("Create", ctx, mock.AnythingOfType("*models.User")).
			Return(nil).
			Times(len(members))

		team, resultMembers, err := s.CreateTeam(ctx, "backend", members)

		require.NoError(t, err)
		assert.Equal(t, "backend", team.Name)
		assert.Len(t, resultMembers, 2)

		teamRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})

	t.Run("ошибка при создании существующей команды", func(t *testing.T) {
		teamRepo.
			On("Exists", ctx, "frontend").
			Return(true, nil).
			Once()

		team, resultMembers, err := s.CreateTeam(ctx, "frontend", members)

		assert.ErrorIs(t, err, domainErrors.ErrTeamExists)
		assert.Nil(t, team)
		assert.Nil(t, resultMembers)

		teamRepo.AssertExpectations(t)
	})
}

func TestTeamService_GetTeam(t *testing.T) {
	ctx := context.Background()

	teamRepo := new(TeamRepositoryMock)
	userRepo := new(UserRepositoryMock)
	s := service.NewTeamService(teamRepo, userRepo)

	t.Run("успешное получение команды", func(t *testing.T) {
		team := &models.Team{Name: "backend"}
		members := []*models.TeamMember{
			{UserID: "user1", Username: "Alice", IsActive: true},
			{UserID: "user2", Username: "Bob", IsActive: false},
		}

		teamRepo.
			On("GetByName", ctx, "backend").
			Return(team, nil).
			Once()

		teamRepo.
			On("GetMembers", ctx, "backend").
			Return(members, nil).
			Once()

		resTeam, resMembers, err := s.GetTeam(ctx, "backend")

		require.NoError(t, err)
		assert.Equal(t, "backend", resTeam.Name)
		assert.Len(t, resMembers, 2)

		teamRepo.AssertExpectations(t)
	})

	t.Run("ошибка при получении несуществующей команды", func(t *testing.T) {
		teamRepo.
			On("GetByName", ctx, "nonexistent").
			Return((*models.Team)(nil), nil).
			Once()

		resTeam, resMembers, err := s.GetTeam(ctx, "nonexistent")

		assert.ErrorIs(t, err, domainErrors.ErrTeamNotFound)
		assert.Nil(t, resTeam)
		assert.Nil(t, resMembers)

		teamRepo.AssertExpectations(t)
	})
}

func TestUserService_SetActive(t *testing.T) {
	ctx := context.Background()

	userRepo := new(UserRepositoryMock)
	s := service.NewUserService(userRepo)

	t.Run("успешная деактивация пользователя", func(t *testing.T) {
		user := &models.User{ID: "user1", IsActive: true}

		userRepo.
			On("GetByID", ctx, "user1").
			Return(user, nil).
			Once()

		userRepo.
			On("SetActive", ctx, "user1", false).
			Return(nil).
			Once()

		userUpdated := &models.User{ID: "user1", IsActive: false}
		userRepo.
			On("GetByID", ctx, "user1").
			Return(userUpdated, nil).
			Once()

		res, err := s.SetActive(ctx, "user1", false)

		require.NoError(t, err)
		assert.False(t, res.IsActive)

		userRepo.AssertExpectations(t)
	})

	t.Run("успешная активация пользователя", func(t *testing.T) {
		user := &models.User{ID: "user2", IsActive: false}

		userRepo.
			On("GetByID", ctx, "user2").
			Return(user, nil).
			Once()

		userRepo.
			On("SetActive", ctx, "user2", true).
			Return(nil).
			Once()

		userUpdated := &models.User{ID: "user2", IsActive: true}
		userRepo.
			On("GetByID", ctx, "user2").
			Return(userUpdated, nil).
			Once()

		res, err := s.SetActive(ctx, "user2", true)

		require.NoError(t, err)
		assert.True(t, res.IsActive)

		userRepo.AssertExpectations(t)
	})

	t.Run("ошибка для несуществующего пользователя", func(t *testing.T) {
		userRepo.
			On("GetByID", ctx, "nonexistent").
			Return((*models.User)(nil), nil).
			Once()

		res, err := s.SetActive(ctx, "nonexistent", true)

		assert.ErrorIs(t, err, domainErrors.ErrUserNotFound)
		assert.Nil(t, res)

		userRepo.AssertExpectations(t)
	})
}

func TestUserService_GetActiveByTeam(t *testing.T) {
	ctx := context.Background()

	userRepo := new(UserRepositoryMock)
	s := service.NewUserService(userRepo)

	t.Run("получение только активных пользователей", func(t *testing.T) {
		users := []*models.User{
			{ID: "user1", IsActive: true},
			{ID: "user3", IsActive: true},
		}

		userRepo.
			On("GetActiveByTeam", ctx, "backend").
			Return(users, nil).
			Once()

		res, err := s.GetActiveByTeam(ctx, "backend")

		require.NoError(t, err)
		assert.Len(t, res, 2)
		for _, u := range res {
			assert.True(t, u.IsActive)
		}

		userRepo.AssertExpectations(t)
	})

	t.Run("пустой список для несуществующей команды", func(t *testing.T) {
		userRepo.
			On("GetActiveByTeam", ctx, "nonexistent").
			Return([]*models.User{}, nil).
			Once()

		res, err := s.GetActiveByTeam(ctx, "nonexistent")

		require.NoError(t, err)
		assert.Empty(t, res)

		userRepo.AssertExpectations(t)
	})
}

func TestUserService_DeactivateTeamAndReassign(t *testing.T) {
	ctx := context.Background()

	userRepo := new(UserRepositoryMock)
	prRepo := new(PRRepositoryMock)

	userService := service.NewUserService(userRepo)
	prService := service.NewPRService(prRepo, userRepo)

	users := []*models.User{
		{ID: "u1", TeamName: "backend", IsActive: true},
		{ID: "u2", TeamName: "backend", IsActive: false},
	}

	userRepo.
		On("GetByTeam", ctx, "backend").
		Return(users, nil).
		Once()

	userRepo.
		On("SetActive", ctx, "u1", false).
		Return(nil).
		Once()

	prRepo.
		On("GetByReviewer", ctx, "u1").
		Return([]*models.PullRequest{}, nil).
		Once()

	err := userService.DeactivateTeamAndReassign(ctx, "backend", prService)

	require.NoError(t, err)
	userRepo.AssertExpectations(t)
	prRepo.AssertExpectations(t)
}

func TestUserService_DeactivateTeamAndReassign_NoActiveUsers(t *testing.T) {
	ctx := context.Background()

	userRepo := new(UserRepositoryMock)
	prRepo := new(PRRepositoryMock)

	userService := service.NewUserService(userRepo)
	prService := service.NewPRService(prRepo, userRepo)

	users := []*models.User{
		{ID: "u1", TeamName: "backend", IsActive: false},
		{ID: "u2", TeamName: "backend", IsActive: false},
	}

	userRepo.
		On("GetByTeam", ctx, "backend").
		Return(users, nil).
		Once()

	err := userService.DeactivateTeamAndReassign(ctx, "backend", prService)

	require.NoError(t, err)
	userRepo.AssertExpectations(t)
	prRepo.AssertExpectations(t)
}

func TestPRService_CreatePR(t *testing.T) {
	ctx := context.Background()

	prRepo := new(PRRepositoryMock)
	userRepo := new(UserRepositoryMock)
	s := service.NewPRService(prRepo, userRepo)

	t.Run("успешное создание PR с назначением ревьюверов", func(t *testing.T) {
		author := &models.User{ID: "user1", TeamName: "backend", IsActive: true}
		candidates := []*models.User{
			{ID: "user2", TeamName: "backend", IsActive: true},
			{ID: "user3", TeamName: "backend", IsActive: true},
			{ID: "user4", TeamName: "backend", IsActive: true},
		}

		prRepo.
			On("Exists", ctx, "pr-1").
			Return(false, nil).
			Once()

		userRepo.
			On("GetByID", ctx, "user1").
			Return(author, nil).
			Once()

		userRepo.
			On("GetActiveByTeam", ctx, "backend").
			Return(candidates, nil).
			Once()

		prRepo.
			On("Create", ctx, mock.AnythingOfType("*models.PullRequest")).
			Return(nil).
			Once()

		prRepo.
			On("AssignReviewers", ctx, "pr-1", mock.AnythingOfType("[]string")).
			Return(nil).
			Once()

		pr, err := s.CreatePR(ctx, "pr-1", "Add feature", "user1")

		require.NoError(t, err)
		assert.Equal(t, "pr-1", pr.ID)
		assert.Equal(t, "Add feature", pr.Name)
		assert.Equal(t, "user1", pr.AuthorID)
		assert.Equal(t, enums.StatusOpen, pr.Status)
		assert.True(t, len(pr.AssignedReviewers) > 0 && len(pr.AssignedReviewers) <= 2)
		for _, r := range pr.AssignedReviewers {
			assert.NotEqual(t, "user1", r)
		}

		prRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})

	t.Run("ошибка при создании существующего PR", func(t *testing.T) {
		prRepo.
			On("Exists", ctx, "pr-2").
			Return(true, nil).
			Once()

		pr, err := s.CreatePR(ctx, "pr-2", "Duplicate", "user2")

		assert.ErrorIs(t, err, domainErrors.ErrPRExists)
		assert.Nil(t, pr)

		prRepo.AssertExpectations(t)
	})

	t.Run("ошибка при несуществующем авторе", func(t *testing.T) {
		prRepo.
			On("Exists", ctx, "pr-3").
			Return(false, nil).
			Once()

		userRepo.
			On("GetByID", ctx, "nonexistent").
			Return((*models.User)(nil), nil).
			Once()

		pr, err := s.CreatePR(ctx, "pr-3", "Feature", "nonexistent")

		assert.ErrorIs(t, err, domainErrors.ErrUserNotFound)
		assert.Nil(t, pr)

		prRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})
}

func TestPRService_MergePR(t *testing.T) {
	ctx := context.Background()

	prRepo := new(PRRepositoryMock)
	userRepo := new(UserRepositoryMock)
	_ = userRepo
	s := service.NewPRService(prRepo, userRepo)

	t.Run("успешный мердж PR", func(t *testing.T) {
		pr := &models.PullRequest{
			ID:     "pr-1",
			Status: enums.StatusOpen,
		}

		prRepo.
			On("GetByID", ctx, "pr-1").
			Return(pr, nil).
			Once()

		prRepo.
			On("Merge", ctx, "pr-1").
			Return(nil).
			Once()

		merged := &models.PullRequest{
			ID:        "pr-1",
			Status:    enums.StatusMerged,
			MergedAt:  func() *time.Time { tt := time.Now(); return &tt }(),
			UpdatedAt: time.Now(),
		}

		prRepo.
			On("GetByID", ctx, "pr-1").
			Return(merged, nil).
			Once()

		res, err := s.MergePR(ctx, "pr-1")

		require.NoError(t, err)
		assert.Equal(t, enums.StatusMerged, res.Status)
		assert.NotNil(t, res.MergedAt)

		prRepo.AssertExpectations(t)
	})

	t.Run("идемпотентность мерджа", func(t *testing.T) {
		merged := &models.PullRequest{
			ID:     "pr-1",
			Status: enums.StatusMerged,
			MergedAt: func() *time.Time {
				tt := time.Now().Add(-time.Hour)
				return &tt
			}(),
		}

		prRepo.
			On("GetByID", ctx, "pr-1").
			Return(merged, nil).
			Once()

		res, err := s.MergePR(ctx, "pr-1")

		require.NoError(t, err)
		assert.Equal(t, enums.StatusMerged, res.Status)
		assert.NotNil(t, res.MergedAt)

		prRepo.AssertExpectations(t)
	})

	t.Run("ошибка при мердже несуществующего PR", func(t *testing.T) {
		prRepo.
			On("GetByID", ctx, "nonexistent").
			Return((*models.PullRequest)(nil), nil).
			Once()

		res, err := s.MergePR(ctx, "nonexistent")

		assert.ErrorIs(t, err, domainErrors.ErrPRNotFound)
		assert.Nil(t, res)

		prRepo.AssertExpectations(t)
	})
}

func TestPRService_ReassignReviewer(t *testing.T) {
	ctx := context.Background()

	prRepo := new(PRRepositoryMock)
	userRepo := new(UserRepositoryMock)
	s := service.NewPRService(prRepo, userRepo)

	t.Run("успешное переназначение ревьювера", func(t *testing.T) {
		pr := &models.PullRequest{
			ID:                "pr-1",
			AuthorID:          "user1",
			Status:            enums.StatusOpen,
			AssignedReviewers: []string{"user2", "user3"},
		}

		oldReviewer := &models.User{ID: "user2", TeamName: "backend", IsActive: true}
		candidates := []*models.User{
			{ID: "user2", TeamName: "backend", IsActive: true},
			{ID: "user3", TeamName: "backend", IsActive: true},
			{ID: "user4", TeamName: "backend", IsActive: true},
		}

		prRepo.
			On("GetByID", ctx, "pr-1").
			Return(pr, nil).
			Once()

		userRepo.
			On("GetByID", ctx, "user2").
			Return(oldReviewer, nil).
			Once()

		userRepo.
			On("GetActiveByTeam", ctx, "backend").
			Return(candidates, nil).
			Once()

		prRepo.
			On("RemoveReviewer", ctx, "pr-1", "user2").
			Return(nil).
			Once()

		prRepo.
			On("AssignReviewers", ctx, "pr-1", mock.AnythingOfType("[]string")).
			Return(nil).
			Once()

		updated := &models.PullRequest{
			ID:                "pr-1",
			AuthorID:          "user1",
			Status:            enums.StatusOpen,
			AssignedReviewers: []string{"user3", "user4"},
		}

		prRepo.
			On("GetByID", ctx, "pr-1").
			Return(updated, nil).
			Once()

		res, newReviewerID, err := s.ReassignReviewer(ctx, "pr-1", "user2")

		require.NoError(t, err)
		assert.NotEqual(t, "user2", newReviewerID)
		assert.NotContains(t, res.AssignedReviewers, "user2")
		assert.Contains(t, res.AssignedReviewers, newReviewerID)

		prRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})

	t.Run("ошибка при переназначении на смерженном PR", func(t *testing.T) {
		pr := &models.PullRequest{
			ID:     "pr-2",
			Status: enums.StatusMerged,
		}

		prRepo.
			On("GetByID", ctx, "pr-2").
			Return(pr, nil).
			Once()

		res, newID, err := s.ReassignReviewer(ctx, "pr-2", "user2")

		assert.ErrorIs(t, err, domainErrors.ErrPRMerged)
		assert.Empty(t, newID)
		assert.Nil(t, res)

		prRepo.AssertExpectations(t)
	})

	t.Run("ошибка при переназначении неназначенного пользователя", func(t *testing.T) {
		pr := &models.PullRequest{
			ID:                "pr-1",
			Status:            enums.StatusOpen,
			AssignedReviewers: []string{"user2"},
		}

		prRepo.
			On("GetByID", ctx, "pr-1").
			Return(pr, nil).
			Once()

		res, newID, err := s.ReassignReviewer(ctx, "pr-1", "user4")

		assert.ErrorIs(t, err, domainErrors.ErrNotAssigned)
		assert.Empty(t, newID)
		assert.Nil(t, res)

		prRepo.AssertExpectations(t)
	})
}

func TestPRService_GetPRsByReviewer(t *testing.T) {
	ctx := context.Background()

	prRepo := new(PRRepositoryMock)
	userRepo := new(UserRepositoryMock)
	s := service.NewPRService(prRepo, userRepo)

	t.Run("получение PR для ревьювера", func(t *testing.T) {
		prs := []*models.PullRequest{
			{ID: "pr-1"},
			{ID: "pr-2"},
		}

		userRepo.
			On("Exists", ctx, "user2").
			Return(true, nil).
			Once()

		prRepo.
			On("GetByReviewer", ctx, "user2").
			Return(prs, nil).
			Once()

		res, err := s.GetPRsByReviewer(ctx, "user2")

		require.NoError(t, err)
		assert.True(t, len(res) >= 0)

		prRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})

	t.Run("ошибка для несуществующего пользователя", func(t *testing.T) {
		userRepo.
			On("Exists", ctx, "nonexistent").
			Return(false, nil).
			Once()

		res, err := s.GetPRsByReviewer(ctx, "nonexistent")

		assert.ErrorIs(t, err, domainErrors.ErrUserNotFound)
		assert.Nil(t, res)

		userRepo.AssertExpectations(t)
	})
}

func TestNewPullRequest(t *testing.T) {
	t.Run("успешное создание PR", func(t *testing.T) {
		pr, err := models.NewPullRequest("pr-1", "Feature X", "user1")

		require.NoError(t, err)
		assert.Equal(t, "pr-1", pr.ID)
		assert.Equal(t, "Feature X", pr.Name)
		assert.Equal(t, "user1", pr.AuthorID)
		assert.Equal(t, enums.StatusOpen, pr.Status)
		assert.Empty(t, pr.AssignedReviewers)
		assert.Nil(t, pr.MergedAt)
	})

	t.Run("ошибка при пустом ID", func(t *testing.T) {
		pr, err := models.NewPullRequest("", "Feature X", "user1")
		assert.Error(t, err)
		assert.Nil(t, pr)
	})

	t.Run("ошибка при пустом имени", func(t *testing.T) {
		pr, err := models.NewPullRequest("pr-1", "", "user1")
		assert.Error(t, err)
		assert.Nil(t, pr)
	})

	t.Run("ошибка при пустом authorID", func(t *testing.T) {
		pr, err := models.NewPullRequest("pr-1", "Feature X", "")
		assert.Error(t, err)
		assert.Nil(t, pr)
	})
}

func TestPullRequest_IsMerged(t *testing.T) {
	pr, _ := models.NewPullRequest("pr-1", "Feature", "user1")
	assert.False(t, pr.IsMerged())

	pr.Status = enums.StatusMerged
	assert.True(t, pr.IsMerged())
}

func TestPullRequest_Merge(t *testing.T) {
	t.Run("мердж открытого PR", func(t *testing.T) {
		pr, _ := models.NewPullRequest("pr-1", "Feature", "user1")
		pr.Merge()

		assert.Equal(t, enums.StatusMerged, pr.Status)
		assert.NotNil(t, pr.MergedAt)
	})

	t.Run("идемпотентность мерджа", func(t *testing.T) {
		pr, _ := models.NewPullRequest("pr-1", "Feature", "user1")
		pr.Merge()
		firstMergedAt := pr.MergedAt

		pr.Merge()
		assert.Equal(t, firstMergedAt, pr.MergedAt)
	})
}

func TestPullRequest_AssignReviewers(t *testing.T) {
	pr, _ := models.NewPullRequest("pr-1", "Feature", "user1")
	reviewers := []string{"user2", "user3"}
	pr.AssignReviewers(reviewers)

	assert.Equal(t, reviewers, pr.AssignedReviewers)
}

func TestPullRequest_IsReviewerAssigned(t *testing.T) {
	pr, _ := models.NewPullRequest("pr-1", "Feature", "user1")
	pr.AssignReviewers([]string{"user2", "user3"})

	assert.True(t, pr.IsReviewerAssigned("user2"))
	assert.True(t, pr.IsReviewerAssigned("user3"))
	assert.False(t, pr.IsReviewerAssigned("user4"))
}

func TestPullRequest_ReplaceReviewer(t *testing.T) {
	t.Run("успешная замена", func(t *testing.T) {
		pr, _ := models.NewPullRequest("pr-1", "Feature", "user1")
		pr.AssignReviewers([]string{"user2", "user3"})

		replaced := pr.ReplaceReviewer("user2", "user4")

		assert.True(t, replaced)
		assert.Contains(t, pr.AssignedReviewers, "user4")
		assert.NotContains(t, pr.AssignedReviewers, "user2")
	})

	t.Run("замена несуществующего ревьювера", func(t *testing.T) {
		pr, _ := models.NewPullRequest("pr-1", "Feature", "user1")
		pr.AssignReviewers([]string{"user2"})

		replaced := pr.ReplaceReviewer("user5", "user6")

		assert.False(t, replaced)
		assert.Equal(t, []string{"user2"}, pr.AssignedReviewers)
	})
}

func TestNewTeam(t *testing.T) {
	t.Run("успешное создание команды", func(t *testing.T) {
		team, err := models.NewTeam("backend")

		require.NoError(t, err)
		assert.Equal(t, "backend", team.Name)
		assert.False(t, team.CreatedAt.IsZero())
		assert.False(t, team.UpdatedAt.IsZero())
	})

	t.Run("ошибка при пустом имени", func(t *testing.T) {
		team, err := models.NewTeam("")

		assert.Error(t, err)
		assert.Nil(t, team)
		assert.Contains(t, err.Error(), "team_name")
	})
}

func TestNewUser(t *testing.T) {
	t.Run("успешное создание пользователя", func(t *testing.T) {
		user, err := models.NewUser("user1", "Alice", "backend", true)

		require.NoError(t, err)
		assert.Equal(t, "user1", user.ID)
		assert.Equal(t, "Alice", user.Username)
		assert.Equal(t, "backend", user.TeamName)
		assert.True(t, user.IsActive)
		assert.False(t, user.CreatedAt.IsZero())
	})

	t.Run("ошибка при пустом ID", func(t *testing.T) {
		user, err := models.NewUser("", "Alice", "backend", true)
		assert.ErrorIs(t, err, models.ErrInvalidUserID)
		assert.Nil(t, user)
	})

	t.Run("ошибка при пустом username", func(t *testing.T) {
		user, err := models.NewUser("user1", "", "backend", true)
		assert.ErrorIs(t, err, models.ErrInvalidUsername)
		assert.Nil(t, user)
	})

	t.Run("ошибка при пустом teamName", func(t *testing.T) {
		user, err := models.NewUser("user1", "Alice", "", true)
		assert.ErrorIs(t, err, models.ErrInvalidTeamName)
		assert.Nil(t, user)
	})
}

func TestUser_SetActive(t *testing.T) {
	t.Run("деактивация пользователя", func(t *testing.T) {
		user, _ := models.NewUser("user1", "Alice", "backend", true)
		oldUpdatedAt := user.UpdatedAt

		time.Sleep(1 * time.Millisecond)
		user.SetActive(false)

		assert.False(t, user.IsActive)
		assert.True(t, user.UpdatedAt.After(oldUpdatedAt))
	})

	t.Run("активация пользователя", func(t *testing.T) {
		user, _ := models.NewUser("user1", "Alice", "backend", false)
		user.SetActive(true)

		assert.True(t, user.IsActive)
	})
}

func TestValidationError_Error(t *testing.T) {
	err := &models.ValidationError{Field: "test_field", Message: "test message"}
	assert.Equal(t, "test_field: test message", err.Error())
}

func TestPRHandler_CreatePR(t *testing.T) {
	ctx := context.Background()

	t.Run("успешное создание PR", func(t *testing.T) {
		prRepo := new(PRRepositoryMock)
		userRepo := new(UserRepositoryMock)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewPRHandler(prService)

		author := &models.User{ID: "user1", TeamName: "backend", IsActive: true}
		candidates := []*models.User{
			{ID: "user2", TeamName: "backend", IsActive: true},
			{ID: "user3", TeamName: "backend", IsActive: true},
		}

		prRepo.On("Exists", ctx, "pr-1").Return(false, nil)
		userRepo.On("GetByID", ctx, "user1").Return(author, nil)
		userRepo.On("GetActiveByTeam", ctx, "backend").Return(candidates, nil)
		prRepo.On("Create", ctx, mock.AnythingOfType("*models.PullRequest")).Return(nil)
		prRepo.On("AssignReviewers", ctx, "pr-1", mock.AnythingOfType("[]string")).Return(nil)

		reqBody := `{"pull_request_id":"pr-1","pull_request_name":"Feature","author_id":"user1"}`
		req := httptest.NewRequest(http.MethodPost, "/api/pull-requests", strings.NewReader(reqBody))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.CreatePR(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		prRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})

	t.Run("ошибка при невалидном JSON", func(t *testing.T) {
		prRepo := new(PRRepositoryMock)
		userRepo := new(UserRepositoryMock)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewPRHandler(prService)

		req := httptest.NewRequest(http.MethodPost, "/api/pull-requests", strings.NewReader("invalid json"))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.CreatePR(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при пустых полях", func(t *testing.T) {
		prRepo := new(PRRepositoryMock)
		userRepo := new(UserRepositoryMock)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewPRHandler(prService)

		reqBody := `{"pull_request_id":"","pull_request_name":"","author_id":""}`
		req := httptest.NewRequest(http.MethodPost, "/api/pull-requests", strings.NewReader(reqBody))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.CreatePR(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при существующем PR", func(t *testing.T) {
		prRepo := new(PRRepositoryMock)
		userRepo := new(UserRepositoryMock)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewPRHandler(prService)

		prRepo.On("Exists", ctx, "pr-1").Return(true, nil)

		reqBody := `{"pull_request_id":"pr-1","pull_request_name":"Feature","author_id":"user1"}`
		req := httptest.NewRequest(http.MethodPost, "/api/pull-requests", strings.NewReader(reqBody))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.CreatePR(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
		prRepo.AssertExpectations(t)
	})
}

func TestPRHandler_MergePR(t *testing.T) {
	ctx := context.Background()

	t.Run("успешный мердж", func(t *testing.T) {
		prRepo := new(PRRepositoryMock)
		userRepo := new(UserRepositoryMock)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewPRHandler(prService)

		pr := &models.PullRequest{ID: "pr-1", Status: enums.StatusOpen}
		merged := &models.PullRequest{
			ID:       "pr-1",
			Status:   enums.StatusMerged,
			MergedAt: func() *time.Time { t := time.Now(); return &t }(),
		}

		prRepo.On("GetByID", ctx, "pr-1").Return(pr, nil).Once()
		prRepo.On("Merge", ctx, "pr-1").Return(nil)
		prRepo.On("GetByID", ctx, "pr-1").Return(merged, nil).Once()

		reqBody := `{"pull_request_id":"pr-1"}`
		req := httptest.NewRequest(http.MethodPost, "/api/pull-requests/merge", strings.NewReader(reqBody))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.MergePR(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		prRepo.AssertExpectations(t)
	})

	t.Run("ошибка при невалидном JSON", func(t *testing.T) {
		prRepo := new(PRRepositoryMock)
		userRepo := new(UserRepositoryMock)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewPRHandler(prService)

		req := httptest.NewRequest(http.MethodPost, "/api/pull-requests/merge", strings.NewReader("bad"))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.MergePR(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при несуществующем PR", func(t *testing.T) {
		prRepo := new(PRRepositoryMock)
		userRepo := new(UserRepositoryMock)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewPRHandler(prService)

		prRepo.On("GetByID", ctx, "pr-999").Return((*models.PullRequest)(nil), nil)

		reqBody := `{"pull_request_id":"pr-999"}`
		req := httptest.NewRequest(http.MethodPost, "/api/pull-requests/merge", strings.NewReader(reqBody))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.MergePR(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		prRepo.AssertExpectations(t)
	})
}

func TestPRHandler_ReassignReviewer(t *testing.T) {
	ctx := context.Background()

	t.Run("успешное переназначение", func(t *testing.T) {
		prRepo := new(PRRepositoryMock)
		userRepo := new(UserRepositoryMock)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewPRHandler(prService)

		pr := &models.PullRequest{
			ID:                "pr-1",
			AuthorID:          "user1",
			Status:            enums.StatusOpen,
			AssignedReviewers: []string{"user2", "user3"},
		}
		oldReviewer := &models.User{ID: "user2", TeamName: "backend"}
		candidates := []*models.User{
			{ID: "user2", TeamName: "backend", IsActive: true},
			{ID: "user3", TeamName: "backend", IsActive: true},
			{ID: "user4", TeamName: "backend", IsActive: true},
		}
		updated := &models.PullRequest{
			ID:                "pr-1",
			Status:            enums.StatusOpen,
			AssignedReviewers: []string{"user3", "user4"},
		}

		prRepo.On("GetByID", ctx, "pr-1").Return(pr, nil).Once()
		userRepo.On("GetByID", ctx, "user2").Return(oldReviewer, nil)
		userRepo.On("GetActiveByTeam", ctx, "backend").Return(candidates, nil)
		prRepo.On("RemoveReviewer", ctx, "pr-1", "user2").Return(nil)
		prRepo.On("AssignReviewers", ctx, "pr-1", mock.AnythingOfType("[]string")).Return(nil)
		prRepo.On("GetByID", ctx, "pr-1").Return(updated, nil).Once()

		reqBody := `{"pull_request_id":"pr-1","old_user_id":"user2"}`
		req := httptest.NewRequest(http.MethodPost, "/api/pull-requests/reassign", strings.NewReader(reqBody))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ReassignReviewer(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		prRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})

	t.Run("ошибка при пустых полях", func(t *testing.T) {
		prRepo := new(PRRepositoryMock)
		userRepo := new(UserRepositoryMock)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewPRHandler(prService)

		reqBody := `{"pull_request_id":"","old_user_id":""}`
		req := httptest.NewRequest(http.MethodPost, "/api/pull-requests/reassign", strings.NewReader(reqBody))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ReassignReviewer(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestTeamHandler_CreateTeam(t *testing.T) {
	ctx := context.Background()

	t.Run("успешное создание команды", func(t *testing.T) {
		teamRepo := new(TeamRepositoryMock)
		userRepo := new(UserRepositoryMock)
		teamService := service.NewTeamService(teamRepo, userRepo)
		handler := handler.NewTeamHandler(teamService)

		teamRepo.On("Exists", ctx, "backend").Return(false, nil)
		teamRepo.On("Create", ctx, mock.AnythingOfType("*models.Team")).Return(nil)
		userRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).Return(nil).Times(2)

		reqBody := `{"team_name":"backend","members":[{"user_id":"u1","username":"Alice","is_active":true},{"user_id":"u2","username":"Bob","is_active":true}]}`
		req := httptest.NewRequest(http.MethodPost, "/api/teams", strings.NewReader(reqBody))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.CreateTeam(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		teamRepo.AssertExpectations(t)
		userRepo.AssertExpectations(t)
	})

	t.Run("ошибка при невалидном JSON", func(t *testing.T) {
		teamRepo := new(TeamRepositoryMock)
		userRepo := new(UserRepositoryMock)
		teamService := service.NewTeamService(teamRepo, userRepo)
		handler := handler.NewTeamHandler(teamService)

		req := httptest.NewRequest(http.MethodPost, "/api/teams", strings.NewReader("bad json"))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.CreateTeam(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при пустом team_name", func(t *testing.T) {
		teamRepo := new(TeamRepositoryMock)
		userRepo := new(UserRepositoryMock)
		teamService := service.NewTeamService(teamRepo, userRepo)
		handler := handler.NewTeamHandler(teamService)

		reqBody := `{"team_name":"","members":[]}`
		req := httptest.NewRequest(http.MethodPost, "/api/teams", strings.NewReader(reqBody))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.CreateTeam(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при существующей команде", func(t *testing.T) {
		teamRepo := new(TeamRepositoryMock)
		userRepo := new(UserRepositoryMock)
		teamService := service.NewTeamService(teamRepo, userRepo)
		handler := handler.NewTeamHandler(teamService)

		teamRepo.On("Exists", ctx, "backend").Return(true, nil)

		reqBody := `{"team_name":"backend","members":[]}`
		req := httptest.NewRequest(http.MethodPost, "/api/teams", strings.NewReader(reqBody))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.CreateTeam(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		teamRepo.AssertExpectations(t)
	})
}

func TestTeamHandler_GetTeam(t *testing.T) {
	ctx := context.Background()

	t.Run("успешное получение команды", func(t *testing.T) {
		teamRepo := new(TeamRepositoryMock)
		userRepo := new(UserRepositoryMock)
		teamService := service.NewTeamService(teamRepo, userRepo)
		handler := handler.NewTeamHandler(teamService)

		team := &models.Team{Name: "backend"}
		members := []*models.TeamMember{
			{UserID: "u1", Username: "Alice", IsActive: true},
		}

		teamRepo.On("GetByName", ctx, "backend").Return(team, nil)
		teamRepo.On("GetMembers", ctx, "backend").Return(members, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/teams?team_name=backend", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetTeam(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		teamRepo.AssertExpectations(t)
	})

	t.Run("ошибка при пустом team_name", func(t *testing.T) {
		teamRepo := new(TeamRepositoryMock)
		userRepo := new(UserRepositoryMock)
		teamService := service.NewTeamService(teamRepo, userRepo)
		handler := handler.NewTeamHandler(teamService)

		req := httptest.NewRequest(http.MethodGet, "/api/teams", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetTeam(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при несуществующей команде", func(t *testing.T) {
		teamRepo := new(TeamRepositoryMock)
		userRepo := new(UserRepositoryMock)
		teamService := service.NewTeamService(teamRepo, userRepo)
		handler := handler.NewTeamHandler(teamService)

		teamRepo.On("GetByName", ctx, "nonexistent").Return((*models.Team)(nil), nil)

		req := httptest.NewRequest(http.MethodGet, "/api/teams?team_name=nonexistent", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetTeam(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		teamRepo.AssertExpectations(t)
	})
}

func TestUserHandler_SetActive(t *testing.T) {
	ctx := context.Background()

	t.Run("успешная активация пользователя", func(t *testing.T) {
		userRepo := new(UserRepositoryMock)
		prRepo := new(PRRepositoryMock)
		userService := service.NewUserService(userRepo)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewUserHandler(userService, prService)

		user := &models.User{ID: "user1", Username: "Alice", TeamName: "backend", IsActive: false}
		updatedUser := &models.User{ID: "user1", Username: "Alice", TeamName: "backend", IsActive: true}

		userRepo.On("GetByID", ctx, "user1").Return(user, nil).Once()
		userRepo.On("SetActive", ctx, "user1", true).Return(nil)
		userRepo.On("GetByID", ctx, "user1").Return(updatedUser, nil).Once()

		reqBody := `{"user_id":"user1","is_active":true}`
		req := httptest.NewRequest(http.MethodPost, "/api/users/set-active", strings.NewReader(reqBody))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.SetActive(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		userRepo.AssertExpectations(t)
	})

	t.Run("ошибка при невалидном JSON", func(t *testing.T) {
		userRepo := new(UserRepositoryMock)
		prRepo := new(PRRepositoryMock)
		userService := service.NewUserService(userRepo)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewUserHandler(userService, prService)

		req := httptest.NewRequest(http.MethodPost, "/api/users/set-active", strings.NewReader("bad"))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.SetActive(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при несуществующем пользователе", func(t *testing.T) {
		userRepo := new(UserRepositoryMock)
		prRepo := new(PRRepositoryMock)
		userService := service.NewUserService(userRepo)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewUserHandler(userService, prService)

		userRepo.On("GetByID", ctx, "user999").Return((*models.User)(nil), nil)

		reqBody := `{"user_id":"user999","is_active":true}`
		req := httptest.NewRequest(http.MethodPost, "/api/users/set-active", strings.NewReader(reqBody))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.SetActive(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		userRepo.AssertExpectations(t)
	})
}

func TestUserHandler_GetReviews(t *testing.T) {
	ctx := context.Background()

	t.Run("успешное получение PR пользователя", func(t *testing.T) {
		userRepo := new(UserRepositoryMock)
		prRepo := new(PRRepositoryMock)
		userService := service.NewUserService(userRepo)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewUserHandler(userService, prService)

		prs := []*models.PullRequest{
			{ID: "pr-1", Name: "Feature A", AuthorID: "author1", Status: enums.StatusOpen},
			{ID: "pr-2", Name: "Feature B", AuthorID: "author2", Status: enums.StatusMerged},
		}

		userRepo.On("Exists", ctx, "user1").Return(true, nil)
		prRepo.On("GetByReviewer", ctx, "user1").Return(prs, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/users/reviews?user_id=user1", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetReviews(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		userRepo.AssertExpectations(t)
		prRepo.AssertExpectations(t)
	})

	t.Run("ошибка при пустом user_id", func(t *testing.T) {
		userRepo := new(UserRepositoryMock)
		prRepo := new(PRRepositoryMock)
		userService := service.NewUserService(userRepo)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewUserHandler(userService, prService)

		req := httptest.NewRequest(http.MethodGet, "/api/users/reviews", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetReviews(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при несуществующем пользователе", func(t *testing.T) {
		userRepo := new(UserRepositoryMock)
		prRepo := new(PRRepositoryMock)
		userService := service.NewUserService(userRepo)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewUserHandler(userService, prService)

		userRepo.On("Exists", ctx, "user999").Return(false, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/users/reviews?user_id=user999", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetReviews(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		userRepo.AssertExpectations(t)
	})
}

func TestUserHandler_DeactivateTeam(t *testing.T) {
	ctx := context.Background()

	t.Run("успешная деактивация команды", func(t *testing.T) {
		userRepo := new(UserRepositoryMock)
		prRepo := new(PRRepositoryMock)
		userService := service.NewUserService(userRepo)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewUserHandler(userService, prService)

		users := []*models.User{
			{ID: "u1", TeamName: "backend", IsActive: true},
		}

		userRepo.On("GetByTeam", ctx, "backend").Return(users, nil)
		userRepo.On("SetActive", ctx, "u1", false).Return(nil)
		prRepo.On("GetByReviewer", ctx, "u1").Return([]*models.PullRequest{}, nil)

		reqBody := `{"teamName":"backend"}`
		req := httptest.NewRequest(http.MethodPost, "/api/users/deactivate-team", strings.NewReader(reqBody))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.DeactivateTeam(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		userRepo.AssertExpectations(t)
		prRepo.AssertExpectations(t)
	})

	t.Run("ошибка при невалидном JSON", func(t *testing.T) {
		userRepo := new(UserRepositoryMock)
		prRepo := new(PRRepositoryMock)
		userService := service.NewUserService(userRepo)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewUserHandler(userService, prService)

		req := httptest.NewRequest(http.MethodPost, "/api/users/deactivate-team", strings.NewReader("bad"))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.DeactivateTeam(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("ошибка при пустом teamName", func(t *testing.T) {
		userRepo := new(UserRepositoryMock)
		prRepo := new(PRRepositoryMock)
		userService := service.NewUserService(userRepo)
		prService := service.NewPRService(prRepo, userRepo)
		handler := handler.NewUserHandler(userService, prService)

		reqBody := `{"teamName":""}`
		req := httptest.NewRequest(http.MethodPost, "/api/users/deactivate-team", strings.NewReader(reqBody))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.DeactivateTeam(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestPRStatus_IsValid(t *testing.T) {
	assert.True(t, enums.StatusOpen.IsValid())
	assert.True(t, enums.StatusMerged.IsValid())
	assert.False(t, enums.PRStatus("invalid").IsValid())
	assert.False(t, enums.PRStatus("").IsValid())
}

func TestPRStatus_String(t *testing.T) {
	assert.Equal(t, "OPEN", enums.StatusOpen.String())
	assert.Equal(t, "MERGED", enums.StatusMerged.String())
}

func TestDomainError_Error(t *testing.T) {
	err := domainErrors.NewDomainError(domainErrors.ErrCodeNotFound, "test not found")
	assert.Contains(t, err.Error(), "test not found")
	assert.Contains(t, err.Error(), "NOT_FOUND")
}

func TestNewDomainError(t *testing.T) {
	err := domainErrors.NewDomainError(domainErrors.ErrCodeTeamExists, "team already exists")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "team already exists")
	assert.Contains(t, err.Error(), "TEAM_EXISTS")
}

func TestUserService_GetByID(t *testing.T) {
	ctx := context.Background()
	userRepo := new(UserRepositoryMock)
	s := service.NewUserService(userRepo)

	t.Run("успешное получение пользователя", func(t *testing.T) {
		user := &models.User{ID: "user1", Username: "Alice", TeamName: "backend"}
		userRepo.On("GetByID", ctx, "user1").Return(user, nil).Once()

		res, err := s.GetByID(ctx, "user1")

		require.NoError(t, err)
		assert.Equal(t, "user1", res.ID)
		assert.Equal(t, "Alice", res.Username)
		userRepo.AssertExpectations(t)
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		userRepo.On("GetByID", ctx, "nonexistent").Return((*models.User)(nil), nil).Once()

		res, err := s.GetByID(ctx, "nonexistent")

		assert.ErrorIs(t, err, domainErrors.ErrUserNotFound)
		assert.Nil(t, res)
		userRepo.AssertExpectations(t)
	})
}
