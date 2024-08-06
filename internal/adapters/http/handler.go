package http

import (
	"encoding/json"
	"github.com/olusolaa/github-monitor/pkg/errors"
	"github.com/olusolaa/github-monitor/pkg/logger"
	"github.com/olusolaa/github-monitor/pkg/pagination"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olusolaa/github-monitor/internal/core/services"
)

func RegisterRoutes(r chi.Router, repoService services.RepositoryService, commitService services.CommitService) {
	r.Route("/api", func(r chi.Router) {
		r.Get("/repos/{owner}/{repo}", getRepository(repoService))
		r.Get("/repos/{owner}/{name}/commits", getCommits(commitService))
		r.Get("/repos/{owner}/{name}/top-authors", getTopCommitAuthors(commitService))
		r.Post("/repos/{owner}/{name}/reset-collection", resetCollection(commitService))
		r.Post("/repos/{owner}/{name}/monitor", monitorRepository(repoService))
	})
}

func monitorRepository(repoService services.RepositoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner := chi.URLParam(r, "owner")
		name := chi.URLParam(r, "name")

		err := repoService.AddRepository(owner, name)
		if err != nil {
			logger.LogError(err)
			errors.HandleError(w, err)
			return
		}

		logger.LogInfo("Repository monitoring triggered for: " + owner + "/" + name)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Repository monitoring triggered successfully"})
	}
}

func getRepository(repoService services.RepositoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner := chi.URLParam(r, "owner")
		repo := chi.URLParam(r, "repo")

		repository, err := repoService.GetRepository(r.Context(), repo, owner)
		if err != nil {
			logger.LogError(err)
			errors.HandleError(w, err)
			return
		}

		logger.LogInfo("Repository details fetched for: " + owner + "/" + repo)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repository)
	}
}

func getCommits(commitService services.CommitService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		owner := chi.URLParam(r, "owner")

		page, pageSize, err := pagination.ParsePaginationParams(r.URL.Query())
		if err != nil {
			logger.LogError(err)
			errors.HandleError(w, err)
			return
		}

		commits, pg, err := commitService.GetCommitsByRepositoryName(r.Context(), owner, name, page, pageSize)
		if err != nil {
			logger.LogError(err)
			errors.HandleError(w, err)
			return
		}

		response := pagination.PagedResponse{
			Pagination: pg,
			Data:       commits,
		}

		logger.LogInfo("Commits fetched for repository: " + owner + "/" + name)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func getTopCommitAuthors(commitService services.CommitService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		owner := chi.URLParam(r, "owner")
		limitStr := r.URL.Query().Get("limit")

		limit := 10 // Default limit
		var err error
		if limitStr != "" {
			limit, err = strconv.Atoi(limitStr)
			if err != nil {
				logger.LogError(err)
				errors.HandleError(w, err)
				return
			}
		}

		authors, err := commitService.GetTopCommitAuthors(r.Context(), owner, name, limit)
		if err != nil {
			logger.LogError(err)
			errors.HandleError(w, err)
			return
		}

		logger.LogInfo("Top commit authors fetched for repository name: " + name)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(authors)
	}
}

func resetCollection(commitService services.CommitService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := chi.URLParam(r, "name")
		owner := chi.URLParam(r, "owner")

		startTimeStr := r.URL.Query().Get("start_time")
		if startTimeStr == "" {
			errMsg := "start_time query parameter is required"
			logger.LogWarning(errMsg)
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}

		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			errMsg := "Invalid start_time format, must be RFC3339"
			logger.LogWarning(errMsg)
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}

		err = commitService.ResetCollection(r.Context(), owner, name, startTime)
		if err != nil {
			logger.LogError(err)
			errors.HandleError(w, err)
			return
		}

		logger.LogInfo("Collection reset successfully for repository name: " + name)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Collection reset successfully"})
	}
}
