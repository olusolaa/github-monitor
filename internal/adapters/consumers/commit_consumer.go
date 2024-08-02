package consumers

import (
	"context"
	"github.com/olusolaa/github-monitor/internal/adapters/queue"
	"github.com/olusolaa/github-monitor/internal/core/services"
	"github.com/olusolaa/github-monitor/pkg/logger"
	"strconv"
)

type CommitConsumer struct {
	consumer          queue.MessageConsumer
	publisher         queue.MessagePublisher
	repositoryService services.RepositoryService
	commitService     services.CommitService
	githubService     services.GitHubService
}

func NewCommitConsumer(consumer queue.MessageConsumer, publisher queue.MessagePublisher, repositoryService services.RepositoryService, commitService services.CommitService, githubService services.GitHubService) *CommitConsumer {
	return &CommitConsumer{
		consumer:          consumer,
		publisher:         publisher,
		repositoryService: repositoryService,
		commitService:     commitService,
		githubService:     githubService,
	}
}

func (cc *CommitConsumer) Start() {
	err := cc.consumer.ConsumeMessages("fetch_commits", cc.handleMessage)
	if err != nil {
		logger.LogError(err)
	}
}

func (cc *CommitConsumer) handleMessage(msg []byte) error {
	repoID := string(msg)

	repoIDInt, _ := strconv.ParseInt(repoID, 10, 64)

	ctx := context.Background()

	owner, name, err := cc.repositoryService.GetOwnerAndRepoName(ctx, repoIDInt)
	if err != nil {
		return err
	}

	commits, err := cc.githubService.FetchCommits(ctx, owner, name, "", "", repoIDInt)
	if err != nil {
		return err
	}

	err = cc.commitService.SaveCommits(ctx, commits)
	if err != nil {
		logger.LogError(err)
		return err
	}

	// Publish an event indicating commits have been processed
	err = cc.publisher.PublishMessage("commits_processed", repoID)
	if err != nil {
		logger.LogError(err)
		return err
	}

	return nil
}
