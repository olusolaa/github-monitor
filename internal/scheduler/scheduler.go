package scheduler

import (
	"context"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/olusolaa/github-monitor/config"
	"github.com/olusolaa/github-monitor/internal/core/services"
	"github.com/olusolaa/github-monitor/pkg/logger"
)

type Scheduler struct {
	monitorService *services.MonitorService
	cfg            *config.Config
}

func NewScheduler(monitorService *services.MonitorService, cfg *config.Config) *Scheduler {
	return &Scheduler{monitorService: monitorService, cfg: cfg}
}

func (s *Scheduler) ScheduleMonitoring(repoID int64) {
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Every(s.cfg.PollInterval).Do(func() {
		s.monitorRepository(repoID)
	})
	scheduler.StartAsync()
}

func (s *Scheduler) monitorRepository(repoID int64) {
	ctx := context.Background()
	if err := s.monitorService.MonitorRepository(ctx, repoID); err != nil {
		logger.LogError(err)
	}
}
