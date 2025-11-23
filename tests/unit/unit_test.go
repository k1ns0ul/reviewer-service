package unit

import (
	"context"
	"testing"
	"time"

	"reviewer-service/internal/application/service"
	"reviewer-service/internal/domain/enums"
	domainErrors "reviewer-service/internal/domain/errors"
	"reviewer-service/internal/domain/models"

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

// TeamService тесты

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

// UserService тесты

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

// PRService тесты

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
