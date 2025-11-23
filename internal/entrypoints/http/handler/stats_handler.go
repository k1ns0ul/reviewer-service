package handler

import (
	"encoding/json"
	"net/http"

	"reviewer-service/internal/infrastructure/repository"
	"reviewer-service/pkg/logger"
)

type StatsHandler struct {
	repo *repository.StatsRepository
}

func NewStatsHandler(repo *repository.StatsRepository) *StatsHandler {
	return &StatsHandler{repo: repo}
}

func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)

	stats, err := h.repo.GetStats(ctx)
	if err != nil {
		log.Error("ошибка получения статистики", "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "internal_error",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(stats)
}
