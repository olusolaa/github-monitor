package container

import (
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/olusolaa/github-monitor/config"
	"github.com/olusolaa/github-monitor/internal/adapters/consumers"
	"github.com/olusolaa/github-monitor/internal/adapters/github"
	"github.com/olusolaa/github-monitor/internal/adapters/postgresdb"
	"github.com/olusolaa/github-monitor/internal/adapters/queue"
	"github.com/olusolaa/github-monitor/internal/core/services"
	"github.com/olusolaa/github-monitor/internal/scheduler"
	"github.com/olusolaa/github-monitor/pkg/httpclient"
	"github.com/olusolaa/github-monitor/pkg/logger"
	"github.com/pkg/errors"
	"net/http"
)

type Container struct {
	cfg                *config.Config
	dbConn             *sqlx.DB
	rabbitMQ           *queue.RabbitMQConnectionManager
	repoService        services.RepositoryService
	commitService      services.CommitService
	monitorService     *services.MonitorService
	gitHubService      services.GitHubService
	publisher          queue.MessagePublisher
	commitConsumer     *consumers.CommitConsumer
	monitoringConsumer *consumers.MonitoringConsumer
}

func NewContainer(cfg *config.Config) *Container {
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:5432/%s?sslmode=disable",
		cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresDB)

	dbConn, err := sqlx.Open("postgres", connStr)
	if err != nil {
		logger.LogError(errors.Wrap(err, "Error connecting to database"))
		panic(err)
	}

	rabbitMQ, err := queue.NewRabbitMQConnectionManager(cfg.RabbitMQURL)
	if err != nil {
		logger.LogError(errors.Wrap(err, "Error connecting to RabbitMQ"))
		panic(err)
	}

	githubRateLimiter := github.NewGitHubRateLimiter()
	ghClient := github.NewClient(cfg.GitHubBaseURL, httpclient.NewClient(http.DefaultClient, githubRateLimiter.RateLimitMiddleware, httpclient.LoggingMiddleware, httpclient.AuthMiddleware(cfg.GitHubToken)))

	repoRepo := postgresdb.NewRepositoryRepository(dbConn)
	commitRepo := postgresdb.NewCommitRepository(dbConn)

	githubService := services.NewGitHubService(ghClient)
	repoService := services.NewRepositoryService(githubService, repoRepo)
	commitService := services.NewCommitService(githubService, repoService, commitRepo)
	monitorService := services.NewMonitorService(repoService, commitService, githubService, cfg.MaxRetries, cfg.InitialBackoff)

	publisher := queue.NewRabbitMQPublisher(rabbitMQ)
	consumer := queue.NewRabbitMQConsumer(rabbitMQ)

	commitConsumer := consumers.NewCommitConsumer(consumer, publisher, repoService, commitService, githubService, cfg)
	sched := scheduler.NewScheduler(monitorService, cfg)
	monitoringConsumer := consumers.NewMonitoringConsumer(consumer, sched)

	return &Container{
		cfg:                cfg,
		dbConn:             dbConn,
		rabbitMQ:           rabbitMQ,
		repoService:        repoService,
		commitService:      commitService,
		gitHubService:      githubService,
		monitorService:     monitorService,
		publisher:          publisher,
		commitConsumer:     commitConsumer,
		monitoringConsumer: monitoringConsumer,
	}
}

func (c *Container) InitializeRepository() {
	err := c.repoService.InitializeRepository(c.publisher, c.cfg.DefaultOwner, c.cfg.DefaultRepo)
	if err != nil {
		panic(fmt.Errorf("error initializing repository: %v", err))
	}
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
