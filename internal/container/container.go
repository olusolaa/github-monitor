package container

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/olusolaa/github-monitor/config"
	"github.com/olusolaa/github-monitor/internal/adapters/consumers"
	"github.com/olusolaa/github-monitor/internal/adapters/github"
	"github.com/olusolaa/github-monitor/internal/adapters/postgresdb"
	"github.com/olusolaa/github-monitor/internal/adapters/queue"
	"github.com/olusolaa/github-monitor/internal/core/initializer"
	"github.com/olusolaa/github-monitor/internal/core/services"
	"github.com/olusolaa/github-monitor/internal/scheduler"
	"github.com/olusolaa/github-monitor/pkg/logger"
)

type Container struct {
	cfg                *config.Config
	dbConn             *sqlx.DB
	rabbitMQ           *queue.RabbitMQConnectionManager
	repoService        services.RepositoryService
	commitService      services.CommitService
	monitorService     *services.MonitorService
	publisher          queue.MessagePublisher
	commitConsumer     *consumers.CommitConsumer
	monitoringConsumer *consumers.MonitoringConsumer
}

func NewContainer(cfg *config.Config) *Container {
	dbConn, err := sqlx.Open("postgres", cfg.DatabaseDSN)
	if err != nil {
		logger.LogError(err)
		panic(err)
	}

	rabbitMQ, err := queue.NewRabbitMQConnectionManager(cfg.RabbitMQURL)
	if err != nil {
		logger.LogError(err)
		panic(err)
	}

	//redisCache := cache.NewRedisCache()
	ghClient := github.NewClient(cfg.GitHubToken)

	repoRepo := postgresdb.NewRepositoryRepository(dbConn)
	commitRepo := postgresdb.NewCommitRepository(dbConn)

	githubService := services.NewGitHubService(ghClient)
	repoService := services.NewRepositoryService(githubService, repoRepo)
	commitService := services.NewCommitService(githubService, repoService, commitRepo)
	monitorService := services.NewMonitorService(commitService, cfg.MaxRetries, cfg.InitialBackoff)

	publisher := queue.NewRabbitMQPublisher(rabbitMQ)
	consumer := queue.NewRabbitMQConsumer(rabbitMQ)

	commitConsumer := consumers.NewCommitConsumer(consumer, publisher, commitService)
	sched := scheduler.NewScheduler(monitorService, cfg)
	monitoringConsumer := consumers.NewMonitoringConsumer(consumer, sched)

	return &Container{
		cfg:                cfg,
		dbConn:             dbConn,
		rabbitMQ:           rabbitMQ,
		repoService:        repoService,
		commitService:      commitService,
		monitorService:     monitorService,
		publisher:          publisher,
		commitConsumer:     commitConsumer,
		monitoringConsumer: monitoringConsumer,
	}
}

func (c *Container) InitializeRepository() {
	initializer.InitializeRepository(c.repoService, c.publisher, c.cfg.DefaultOwner, c.cfg.DefaultRepo)
}

func (c *Container) StartCommitConsumer() {
	go c.commitConsumer.Start()
}

func (c *Container) StartMonitoringConsumer() {
	go c.monitoringConsumer.Start()
}

func (c *Container) GetRepoService() services.RepositoryService {
	return c.repoService
}

func (c *Container) GetCommitService() services.CommitService {
	return c.commitService
}

func (c *Container) Close() {
	c.dbConn.Close()
	c.rabbitMQ.Close()
}
