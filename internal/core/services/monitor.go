package services

import (
	"context"
	"time"

	"github.com/olusolaa/github-monitor/pkg/logger"
	"github.com/olusolaa/github-monitor/pkg/utils"
)

type MonitorService struct {
	commitService  CommitService
	maxRetries     int
	initialBackoff time.Duration
}

func NewMonitorService(commitService CommitService, maxRetries int, initialBackoff time.Duration) *MonitorService {
	return &MonitorService{commitService: commitService, maxRetries: maxRetries, initialBackoff: initialBackoff}
}

func (m *MonitorService) MonitorRepository(ctx context.Context, repoID int64) error {
	retryCount := 0
	for {
		err := m.monitor(ctx, repoID)
		if err == nil {
			break
		}

		logger.LogError(err)
		retryCount++
		if retryCount >= m.maxRetries {
			return err
		}

		backoffDuration := utils.ExponentialBackoff(retryCount, m.initialBackoff)
		time.Sleep(backoffDuration)
	}
	return nil
}

func (m *MonitorService) monitor(ctx context.Context, repoID int64) error {
	latestCommit, err := m.commitService.GetLatestCommit(ctx, repoID)
	if err != nil {
		return err
	}

	var since string
	if latestCommit != nil {
		since = latestCommit.CommitDate.Format(time.RFC3339)
	}

	commits, err := m.commitService.FetchCommits(ctx, repoID, since, "")
	if err != nil {
		return err
	}

	return m.commitService.SaveCommits(ctx, commits)
}
