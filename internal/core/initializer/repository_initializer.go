package initializer

import (
	"context"
	"strconv"

	"github.com/olusolaa/github-monitor/internal/adapters/queue"
	"github.com/olusolaa/github-monitor/internal/core/services"
	"github.com/olusolaa/github-monitor/pkg/logger"
)

func InitializeRepository(repoService services.RepositoryService, publisher queue.MessagePublisher, owner, repo string) {
	ctx := context.Background()

	repository, err := repoService.FetchRepositoryInfo(ctx, repo, owner)
	if err != nil {
		logger.LogError(err)
		return
	}

	err = publisher.PublishMessage("fetch_commits", strconv.Itoa(int(repository.ID)))
	if err != nil {
		logger.LogError(err)
		return
	}

	logger.LogInfo("Initialized repository and published event for fetching commits")
}
