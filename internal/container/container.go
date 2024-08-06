package container

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"net/http"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/olusolaa/github-monitor/config"
	"github.com/olusolaa/github-monitor/internal/adapters/github"
	"github.com/olusolaa/github-monitor/internal/adapters/postgresdb"
	"github.com/olusolaa/github-monitor/internal/core/services"
	"github.com/olusolaa/github-monitor/internal/scheduler"
	"github.com/olusolaa/github-monitor/pkg/httpclient"
	"github.com/pkg/errors"
)

type Container struct {
	cfg            *config.Config
	dbConn         *sqlx.DB
	repoService    services.RepositoryService
	commitService  services.CommitService
	monitorService *services.MonitorService
	gitHubService  services.GitHubService
	scheduler      *scheduler.Scheduler
	commitChan     chan int64
	monitoringChan chan int64
}

func NewContainer(cfg *config.Config) *Container {
	connStr := fmt.Sprintf("postgresql://%s:%s@%s:5432/%s?sslmode=disable",
		cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresDB)

	dbConn, err := initializeDatabase(connStr)
	if err != nil {
		panic(errors.Wrap(err, "Error connecting to database"))
	}

	if err := runMigrations(connStr); err != nil {
		panic(errors.Wrap(err, "Error running migrations"))
	}

	githubRateLimiter := github.NewGitHubRateLimiter()
	ghClient := github.NewClient(cfg.GitHubBaseURL, httpclient.NewClient(http.DefaultClient, githubRateLimiter.RateLimitMiddleware, httpclient.LoggingMiddleware, httpclient.AuthMiddleware(cfg.GitHubToken)))

	repoRepo := postgresdb.NewRepositoryRepository(dbConn)
	commitRepo := postgresdb.NewCommitRepository(dbConn)

	githubService := services.NewGitHubService(ghClient)
	commitChan := make(chan int64, 100)     // Initialize commitChan with a buffer size
	monitoringChan := make(chan int64, 100) // Initialize monitoringChan with a buffer size

	repoChan := make(chan services.RepoRequest, 10) // Buffered channel for concurrent requests
	repoService := services.NewRepositoryService(githubService, repoRepo, repoChan, commitChan)
	commitService := services.NewCommitService(githubService, repoService, commitRepo, commitChan)
	monitorService := services.NewMonitorService(repoService, commitService, githubService, cfg.MaxRetries, cfg.InitialBackoff)
	schedulerService := scheduler.NewScheduler(monitorService, cfg)

	return &Container{
		cfg:            cfg,
		dbConn:         dbConn,
		repoService:    repoService,
		commitService:  commitService,
		gitHubService:  githubService,
		monitorService: monitorService,
		scheduler:      schedulerService,
		commitChan:     commitChan,
		monitoringChan: monitoringChan,
	}
}

func initializeDatabase(connStr string) (*sqlx.DB, error) {

	dbConn, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := dbConn.Ping(); err != nil {
		return nil, err
	}

	return dbConn, nil
}

func (c *Container) InitializeRepository() {
	err := c.repoService.AddRepository(c.cfg.DefaultOwner, c.cfg.DefaultRepo)
	if err != nil {
		panic(fmt.Errorf("error initializing repository: %v", err))
	}
}

func (c *Container) GetRepoService() services.RepositoryService {
	return c.repoService
}

func (c *Container) GetCommitService() services.CommitService {
	return c.commitService
}

func (c *Container) StartServices() {
	go c.commitService.CommitManager(c.monitoringChan, c.cfg.StartDate, c.cfg.EndDate)
	go c.scheduler.ScheduleMonitoring(c.monitoringChan)
}

func (c *Container) Close() {
	c.dbConn.Close()
}

func runMigrations(dsn string) error {
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
