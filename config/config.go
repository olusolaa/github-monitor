package config

import (
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ServerAddress  string
	GitHubToken    string
	DatabaseDSN    string
	PollInterval   time.Duration
	MaxRetries     int
	InitialBackoff time.Duration
	StartDate      string
	WebhookSecret  string
	LogLevel       string
	DefaultOwner   string
	DefaultRepo    string
	RedisURL       string
	RabbitMQURL    string
}

func LoadConfig() *Config {
	viper.SetConfigFile(".env")

	// Allow reading from environment variables
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("SERVER_ADDRESS", "0.0.0.0:8080")
	viper.SetDefault("POLL_INTERVAL", 3600) // 1 hour in seconds
	viper.SetDefault("MAX_RETRIES", 3)
	viper.SetDefault("INITIAL_BACKOFF", 2) // In seconds
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("REDIS_URL", "redis://redis:6379")                   // Docker service default
	viper.SetDefault("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/") // Docker service default
	viper.SetDefault("GITHUB_TOKEN", "default_github_token")
	viper.SetDefault("DATABASE_DSN", "postgresql://postgres:password@postgres:5432/postgres?sslmode=disable")
	viper.SetDefault("START_DATE", "2023-01-01T00:00:00Z")
	viper.SetDefault("WEBHOOK_SECRET", "default_webhook_secret")
	viper.SetDefault("DEFAULT_OWNER", "chromium")
	viper.SetDefault("DEFAULT_REPO", "chromium")

	// Read in .env file, if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// No .env file found, continue with environment variables and defaults
			log.Println("No .env file found, continuing with environment variables and defaults")
		} else {
			// Some other error occurred while reading the config file
			log.Fatalf("Error reading .env file: %v", err)
		}
	}

	return &Config{
		ServerAddress:  viper.GetString("SERVER_ADDRESS"),
		GitHubToken:    viper.GetString("GITHUB_TOKEN"),
		DatabaseDSN:    viper.GetString("DATABASE_DSN"),
		PollInterval:   time.Duration(viper.GetInt("POLL_INTERVAL")) * time.Second,
		MaxRetries:     viper.GetInt("MAX_RETRIES"),
		InitialBackoff: time.Duration(viper.GetInt("INITIAL_BACKOFF")) * time.Second,
		StartDate:      viper.GetString("START_DATE"),
		WebhookSecret:  viper.GetString("WEBHOOK_SECRET"),
		LogLevel:       viper.GetString("LOG_LEVEL"),
		DefaultOwner:   viper.GetString("DEFAULT_OWNER"),
		DefaultRepo:    viper.GetString("DEFAULT_REPO"),
		RedisURL:       viper.GetString("REDIS_URL"),
		RabbitMQURL:    viper.GetString("RABBITMQ_URL"),
	}
}
