package http

import (
	"encoding/json"
	"github.com/olusolaa/github-monitor/pkg/errors"
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
		r.Get("/repos/{repo_id}/top-authors", getTopCommitAuthors(commitService))
		r.Post("/repos/{repo_id}/reset-collection", resetCollection(commitService))
		r.Post("/repos/{owner}/{name}/monitor", monitorRepository(repoService))
	})
}

func monitorRepository(repoService services.RepositoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner := chi.URLParam(r, "owner")
		name := chi.URLParam(r, "name")

		err := repoService.AddRepository(r.Context(), owner, name)
		if err != nil {
			errors.HandleError(w, err)
			return
		}

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
			errors.HandleError(w, err)
			return
		}

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
			errors.HandleError(w, err)
			return
		}

		commits, pg, err := commitService.GetCommitsByRepositoryName(r.Context(), owner, name, page, pageSize)
		if err != nil {
			errors.HandleError(w, err)
			return
		}

		response := pagination.PagedResponse{
			Pagination: pg,
			Data:       commits,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func getTopCommitAuthors(commitService services.CommitService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repoID := chi.URLParam(r, "repo_id")
		limitStr := r.URL.Query().Get("limit")

		repoIDInt, err := strconv.ParseInt(repoID, 10, 64)
		if err != nil {
			errors.HandleError(w, err)
			return
		}

		limit := 10 // Default limit
		if limitStr != "" {
			limit, err = strconv.Atoi(limitStr)
			if err != nil {
				errors.HandleError(w, err)
				return
			}
		}

		authors, err := commitService.GetTopCommitAuthors(r.Context(), repoIDInt, limit)
		if err != nil {
			errors.HandleError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(authors)
	}
}

func resetCollection(commitService services.CommitService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repoID := chi.URLParam(r, "repo_id")
		repoIDInt, err := strconv.ParseInt(repoID, 10, 64)
		if err != nil {
			errors.HandleError(w, err)
			return
		}

		startTimeStr := r.URL.Query().Get("start_time")
		if startTimeStr == "" {
			http.Error(w, "start_time query parameter is required", http.StatusBadRequest)
			return
		}

		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			http.Error(w, "Invalid start_time format, must be RFC3339", http.StatusBadRequest)
			return
		}

		err = commitService.ResetCollection(r.Context(), repoIDInt, startTime)
		if err != nil {
			errors.HandleError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Collection reset successfully"})
	}
}
