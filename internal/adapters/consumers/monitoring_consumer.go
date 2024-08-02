package consumers

import (
	"github.com/olusolaa/github-monitor/internal/adapters/queue"
	"github.com/olusolaa/github-monitor/internal/scheduler"
	"github.com/olusolaa/github-monitor/pkg/logger"
	"strconv"
)

type MonitoringConsumer struct {
	consumer  queue.MessageConsumer
	scheduler *scheduler.Scheduler
}

func NewMonitoringConsumer(consumer queue.MessageConsumer, scheduler *scheduler.Scheduler) *MonitoringConsumer {
	return &MonitoringConsumer{
		consumer:  consumer,
		scheduler: scheduler,
	}
}

func (mc *MonitoringConsumer) Start() {
	err := mc.consumer.ConsumeMessages("commits_processed", mc.handleMessage)
	if err != nil {
		logger.LogError(err)
	}
}

func (mc *MonitoringConsumer) handleMessage(msg []byte) error {
	repoID := string(msg)
	repoIDInt, _ := strconv.ParseInt(repoID, 10, 64)
	mc.scheduler.ScheduleMonitoring(repoIDInt)
	return nil
}
