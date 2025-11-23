package handler

import (
	"encoding/json"
	"net/http"

	"reviewer-service/internal/application/service"
	"reviewer-service/internal/domain/errors"
	"reviewer-service/internal/entrypoints/http/dto"
	"reviewer-service/pkg/logger"
)

type UserHandler struct {
	userService *service.UserService
	prService   *service.PRService
}

func NewUserHandler(userService *service.UserService, prService *service.PRService) *UserHandler {
	return &UserHandler{
		userService: userService,
		prService:   prService,
	}
}

func (h *UserHandler) SetActive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	var req dto.SetActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("ошибка декодирования запроса", "error", err)
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	user, err := h.userService.SetActive(ctx, req.UserID, req.IsActive)
	if err != nil {
		if err == errors.ErrUserNotFound {
			respondWithError(w, http.StatusNotFound, string(errors.ErrCodeNotFound), err.Error())
			return
		}
		log.Error("ошибка обновления статуса", "error", err)
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update user")
		return
	}

	response := dto.UserResponse{
		User: dto.UserData{
			UserID:   user.ID,
			Username: user.Username,
			TeamName: user.TeamName,
			IsActive: user.IsActive,
		},
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (h *UserHandler) GetReviews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "user_id is required")
		return
	}

	prs, err := h.prService.GetPRsByReviewer(ctx, userID)
	if err != nil {
		if err == errors.ErrUserNotFound {
			respondWithError(w, http.StatusNotFound, string(errors.ErrCodeNotFound), err.Error())
			return
		}
		log.Error("ошибка получения PR", "error", err)
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get reviews")
		return
	}

	pullRequests := make([]dto.PullRequestShort, 0, len(prs))
	for _, pr := range prs {
		pullRequests = append(pullRequests, dto.PullRequestShort{
			PullRequestID:   pr.ID,
			PullRequestName: pr.Name,
			AuthorID:        pr.AuthorID,
			Status:          string(pr.Status),
		})
	}

	response := dto.UserReviewsResponse{
		UserID:       userID,
		PullRequests: pullRequests,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (h *UserHandler) DeactivateTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	var req dto.DeactivateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("ошибка декодирования запроса деактивации команды", "error", err)
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.TeamName == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "teamName is required")
		return
	}

	if err := h.userService.DeactivateTeamAndReassign(ctx, req.TeamName, h.prService); err != nil {
		log.Error("ошибка массовой деактивации команды", "error", err, "team_name", req.TeamName)
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to deactivate team")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"status":   "ok",
		"teamName": req.TeamName,
	})
}
