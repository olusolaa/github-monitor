package consumers

import (
	"context"
	"fmt"
	"github.com/olusolaa/github-monitor/config"
	"github.com/olusolaa/github-monitor/internal/adapters/queue"
	"github.com/olusolaa/github-monitor/internal/core/domain"
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
	cfg               *config.Config
}

func NewCommitConsumer(consumer queue.MessageConsumer, publisher queue.MessagePublisher, repositoryService services.RepositoryService, commitService services.CommitService, githubService services.GitHubService, cfg *config.Config) *CommitConsumer {
	return &CommitConsumer{
		consumer:          consumer,
		publisher:         publisher,
		repositoryService: repositoryService,
		commitService:     commitService,
		githubService:     githubService,
		cfg:               cfg,
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

	domainCommitsChan := make(chan []domain.Commit)
	errChan := make(chan error)

	go cc.githubService.FetchCommits(ctx, owner, name, cc.cfg.StartDate, cc.cfg.EndDate, repoIDInt, domainCommitsChan, errChan)

	var fetchError error

	for {
		select {
		case domainCommits, ok := <-domainCommitsChan:
			if !ok {
				domainCommitsChan = nil
			} else {
				if err := cc.commitService.SaveCommits(ctx, domainCommits); err != nil {
					logger.LogError(err)
					if fetchError == nil {
						fetchError = err
					}
				}
			}
		case err, ok := <-errChan:
			if !ok {
				errChan = nil
			} else if err != nil && fetchError == nil {
				fetchError = fmt.Errorf("error fetching commits: %w", err)
			}
		case <-ctx.Done():
			if fetchError == nil {
				fetchError = ctx.Err()
			}
		}

		if domainCommitsChan == nil && errChan == nil {
			break
		}
	}

	// Publish an event indicating commits have been processed
	if err := cc.publisher.PublishMessage("commits_processed", repoID); err != nil {
		logger.LogError(err)
		return err
	}

	return fetchError
}
