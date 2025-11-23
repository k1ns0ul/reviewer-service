package handler

import (
	"encoding/json"
	"net/http"

	"reviewer-service/internal/application/service"
	"reviewer-service/internal/domain/errors"
	"reviewer-service/internal/domain/models"
	"reviewer-service/internal/entrypoints/http/dto"
	"reviewer-service/pkg/logger"
)

type TeamHandler struct {
	teamService *service.TeamService
}

func NewTeamHandler(teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{
		teamService: teamService,
	}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	var req dto.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("ошибка декодирования запроса", "error", err)
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.TeamName == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "team_name is required")
		return
	}

	members := make([]*models.TeamMember, 0, len(req.Members))
	for _, m := range req.Members {
		members = append(members, &models.TeamMember{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	_, resultMembers, err := h.teamService.CreateTeam(ctx, req.TeamName, members)
	if err != nil {
		if err == errors.ErrTeamExists {
			respondWithError(w, http.StatusBadRequest, string(errors.ErrCodeTeamExists), err.Error())
			return
		}
		log.Error("ошибка создания команды", "error", err)
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create team")
		return
	}

	dtoMembers := make([]dto.TeamMember, 0, len(resultMembers))
	for _, m := range resultMembers {
		dtoMembers = append(dtoMembers, dto.TeamMember{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	response := dto.TeamResponse{
		Team: dto.TeamData{
			TeamName: req.TeamName,
			Members:  dtoMembers,
		},
	}

	respondWithJSON(w, http.StatusCreated, response)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "team_name is required")
		return
	}

	team, members, err := h.teamService.GetTeam(ctx, teamName)
	if err != nil {
		if err == errors.ErrTeamNotFound {
			respondWithError(w, http.StatusNotFound, string(errors.ErrCodeNotFound), err.Error())
			return
		}
		log.Error("ошибка получения команды", "error", err)
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get team")
		return
	}

	dtoMembers := make([]dto.TeamMember, 0, len(members))
	for _, m := range members {
		dtoMembers = append(dtoMembers, dto.TeamMember{
			UserID:   m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	response := dto.TeamData{
		TeamName: team.Name,
		Members:  dtoMembers,
	}

	respondWithJSON(w, http.StatusOK, response)
}
