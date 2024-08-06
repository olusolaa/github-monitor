package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/olusolaa/github-monitor/config"
	"github.com/olusolaa/github-monitor/internal/core/services"
	"github.com/olusolaa/github-monitor/pkg/logger"
)

type Scheduler struct {
	monitorService *services.MonitorService
	cfg            *config.Config
	schedulers     map[int64]*gocron.Scheduler // Map to track schedulers by repo ID
}

func NewScheduler(monitorService *services.MonitorService, cfg *config.Config) *Scheduler {
	return &Scheduler{
		monitorService: monitorService,
		cfg:            cfg,
		schedulers:     make(map[int64]*gocron.Scheduler),
	}
}

func (s *Scheduler) ScheduleMonitoring(monitoringChan chan int64) {
	for repoID := range monitoringChan {
		logger.LogInfo(fmt.Sprintf("Monitoring scheduled for repository ID: %d", repoID))
		if _, exists := s.schedulers[repoID]; !exists {
			scheduler := gocron.NewScheduler(time.UTC)
			s.schedulerJob(scheduler, repoID)
			s.schedulerStart(scheduler, repoID)
		}
	}
}

func (s *Scheduler) schedulerJob(scheduler *gocron.Scheduler, repoID int64) {
	scheduler.Every(s.cfg.PollInterval).Do(func() {
		s.monitorRepository(repoID)
	})
}

func (s *Scheduler) schedulerStart(scheduler *gocron.Scheduler, repoID int64) {
	s.schedulers[repoID] = scheduler
	scheduler.StartAsync()
}

func (s *Scheduler) monitorRepository(repoID int64) {
	ctx := context.Background()
	if err := s.monitorService.MonitorRepository(ctx, repoID); err != nil {
		logger.LogError(fmt.Errorf("monitoring failed for repository ID %d: %w", repoID, err))
	}
}
