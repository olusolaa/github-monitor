package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/olusolaa/github-monitor/internal/core/services"
)

func RegisterRoutes(r chi.Router, repoService services.RepositoryService, commitService services.CommitService) {
	//r.Use(APIKeyAuthMiddleware)
	r.Route("/api", func(r chi.Router) {
		r.Get("/repos/{owner}/{repo}", getRepository(repoService))
		r.Get("/repos/{repo_id}/commits", getCommits(commitService))
		r.Get("/repos/{repo_id}/top-authors", getTopCommitAuthors(commitService))
		r.Post("/repos/{repo_id}/reset-collection", resetCollection(commitService))
	})
}

func getRepository(repoService services.RepositoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner := chi.URLParam(r, "owner")
		repo := chi.URLParam(r, "repo")

		// Retrieve repository details
		repository, err := repoService.FetchRepositoryInfo(r.Context(), repo, owner)
		if err != nil {
			handleError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repository)
	}
}

func getCommits(commitService services.CommitService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repoID := chi.URLParam(r, "repo_id")
		repoIDInt, err := strconv.ParseInt(repoID, 10, 64)

		commits, err := commitService.GetCommitsByRepository(r.Context(), repoIDInt)
		if err != nil {
			handleError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(commits)
	}
}

func getTopCommitAuthors(commitService services.CommitService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repoID := chi.URLParam(r, "repo_id")
		limitStr := r.URL.Query().Get("limit")

		repoIDInt, err := strconv.ParseInt(repoID, 10, 64)
		if err != nil {
			handleError(w, err)
			return
		}

		limit := 10 // Default limit
		if limitStr != "" {
			limit, err = strconv.Atoi(limitStr)
			if err != nil {
				handleError(w, err)
				return
			}
		}

		authors, err := commitService.GetTopCommitAuthors(r.Context(), repoIDInt, limit)
		if err != nil {
			handleError(w, err)
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
			handleError(w, err)
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
			handleError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Collection reset successfully"})
	}
}

func handleError(w http.ResponseWriter, err error) {
	var status int
	var message string

	switch {
	//case errors.Is(err, services.ErrNotFound):
	//	status = http.StatusNotFound
	//	message = "Resource not found"
	//case errors.Is(err, services.ErrInvalidRequest):
	//	status = http.StatusBadRequest
	//	message = "Invalid request"
	default:
		status = http.StatusInternalServerError
		message = "Internal server error"
	}

	http.Error(w, message, status)
}
