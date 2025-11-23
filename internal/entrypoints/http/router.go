package http

import (
	"log/slog"
	"net/http"

	"reviewer-service/internal/application/service"
	"reviewer-service/internal/entrypoints/http/handler"
	"reviewer-service/internal/entrypoints/http/middleware"
	"reviewer-service/internal/infrastructure/repository"

	"github.com/go-chi/chi/v5"
)

func NewRouter(
	log *slog.Logger,
	userService *service.UserService,
	teamService *service.TeamService,
	prService *service.PRService,
	statsRepo *repository.StatsRepository,
) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recovery)
	r.Use(middleware.Logger(log))

	userHandler := handler.NewUserHandler(userService, prService)
	teamHandler := handler.NewTeamHandler(teamService)
	prHandler := handler.NewPRHandler(prService)
	statsHandler := handler.NewStatsHandler(statsRepo)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Get("/stats", statsHandler.GetStats)

	r.Post("/team/add", teamHandler.CreateTeam)
	r.Get("/team/get", teamHandler.GetTeam)
	r.Post("/team/deactivate", userHandler.DeactivateTeam)

	r.Post("/users/setIsActive", userHandler.SetActive)
	r.Get("/users/getReview", userHandler.GetReviews)

	r.Post("/pullRequest/create", prHandler.CreatePR)
	r.Post("/pullRequest/merge", prHandler.MergePR)
	r.Post("/pullRequest/reassign", prHandler.ReassignReviewer)

	return r
}
