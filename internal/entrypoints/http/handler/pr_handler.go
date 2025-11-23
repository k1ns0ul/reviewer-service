package handler

import (
	"encoding/json"
	"net/http"

	"reviewer-service/internal/application/service"
	"reviewer-service/internal/domain/errors"
	"reviewer-service/internal/entrypoints/http/dto"
	"reviewer-service/pkg/logger"
)

type PRHandler struct {
	prService *service.PRService
}

func NewPRHandler(prService *service.PRService) *PRHandler {
	return &PRHandler{
		prService: prService,
	}
}

func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	var req dto.CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("ошибка декодирования запроса", "error", err)
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	// Валидация
	if req.PullRequestID == "" || req.PullRequestName == "" || req.AuthorID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "missing required fields")
		return
	}

	pr, err := h.prService.CreatePR(ctx, req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		if err == errors.ErrPRExists {
			respondWithError(w, http.StatusConflict, string(errors.ErrCodePRExists), err.Error())
			return
		}
		if err == errors.ErrUserNotFound {
			respondWithError(w, http.StatusNotFound, string(errors.ErrCodeNotFound), "author not found")
			return
		}
		log.Error("ошибка создания PR", "error", err)
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create PR")
		return
	}

	response := dto.PRResponse{
		PR: dto.PullRequestData{
			PullRequestID:     pr.ID,
			PullRequestName:   pr.Name,
			AuthorID:          pr.AuthorID,
			Status:            string(pr.Status),
			AssignedReviewers: pr.AssignedReviewers,
			CreatedAt:         &pr.CreatedAt,
		},
	}

	respondWithJSON(w, http.StatusCreated, response)
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	var req dto.MergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("ошибка декодирования запроса", "error", err)
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.PullRequestID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "pull_request_id is required")
		return
	}

	pr, err := h.prService.MergePR(ctx, req.PullRequestID)
	if err != nil {
		if err == errors.ErrPRNotFound {
			respondWithError(w, http.StatusNotFound, string(errors.ErrCodeNotFound), err.Error())
			return
		}
		log.Error("ошибка мерджа PR", "error", err)
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to merge PR")
		return
	}

	response := dto.PRResponse{
		PR: dto.PullRequestData{
			PullRequestID:     pr.ID,
			PullRequestName:   pr.Name,
			AuthorID:          pr.AuthorID,
			Status:            string(pr.Status),
			AssignedReviewers: pr.AssignedReviewers,
			MergedAt:          pr.MergedAt,
		},
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	var req dto.ReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("ошибка декодирования запроса", "error", err)
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if req.PullRequestID == "" || req.OldUserID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "missing required fields")
		return
	}

	pr, newReviewerID, err := h.prService.ReassignReviewer(ctx, req.PullRequestID, req.OldUserID)
	if err != nil {
		if err == errors.ErrPRNotFound || err == errors.ErrUserNotFound {
			respondWithError(w, http.StatusNotFound, string(errors.ErrCodeNotFound), err.Error())
			return
		}
		if err == errors.ErrPRMerged || err == errors.ErrNotAssigned || err == errors.ErrNoCandidate {
			code := string(errors.ErrCodePRMerged)
			if err == errors.ErrNotAssigned {
				code = string(errors.ErrCodeNotAssigned)
			} else if err == errors.ErrNoCandidate {
				code = string(errors.ErrCodeNoCandidate)
			}
			respondWithError(w, http.StatusConflict, code, err.Error())
			return
		}
		log.Error("ошибка переназначения ревьювера", "error", err)
		respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to reassign reviewer")
		return
	}

	response := dto.ReassignResponse{
		PR: dto.PullRequestData{
			PullRequestID:     pr.ID,
			PullRequestName:   pr.Name,
			AuthorID:          pr.AuthorID,
			Status:            string(pr.Status),
			AssignedReviewers: pr.AssignedReviewers,
		},
		ReplacedBy: newReviewerID,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, status int, code, message string) {
	respondWithJSON(w, status, dto.NewErrorResponse(code, message))
}
