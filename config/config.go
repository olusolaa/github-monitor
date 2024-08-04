package config

import (
	"errors"
	"log"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	ServerAddress    string
	GitHubToken      string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresHost     string
	PollInterval     time.Duration
	MaxRetries       int
	InitialBackoff   time.Duration
	StartDate        string
	EndDate          string
	WebhookSecret    string
	LogLevel         string
	DefaultOwner     string
	DefaultRepo      string
	RabbitMQURL      string
	GitHubBaseURL    string
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
	viper.SetDefault("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/") // Docker service default
	viper.SetDefault("GITHUB_TOKEN", "default_github_token")
	viper.SetDefault("START_DATE", "2024-08-03T15:20:00Z")
	viper.SetDefault("END_DATE", "2024-08-03T15:30:00Z")
	viper.SetDefault("WEBHOOK_SECRET", "default_webhook_secret")
	viper.SetDefault("DEFAULT_OWNER", "chromium")
	viper.SetDefault("DEFAULT_REPO", "chromium")
	viper.SetDefault("GITHUB_BASE_URL", "https://api.github.com")
	viper.SetDefault("POSTGRES_USER", "postgres")
	viper.SetDefault("POSTGRES_PASSWORD", "password")
	viper.SetDefault("POSTGRES_DB", "postgres")

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Println("No .env file found, continuing with environment variables and defaults")
		}
	}

	return &Config{
		ServerAddress:    viper.GetString("SERVER_ADDRESS"),
		GitHubToken:      viper.GetString("GITHUB_TOKEN"),
		PollInterval:     time.Duration(viper.GetInt("POLL_INTERVAL")) * time.Second,
		MaxRetries:       viper.GetInt("MAX_RETRIES"),
		InitialBackoff:   time.Duration(viper.GetInt("INITIAL_BACKOFF")) * time.Second,
		StartDate:        viper.GetString("START_DATE"),
		EndDate:          viper.GetString("END_DATE"),
		WebhookSecret:    viper.GetString("WEBHOOK_SECRET"),
		LogLevel:         viper.GetString("LOG_LEVEL"),
		DefaultOwner:     viper.GetString("DEFAULT_OWNER"),
		DefaultRepo:      viper.GetString("DEFAULT_REPO"),
		RabbitMQURL:      viper.GetString("RABBITMQ_URL"),
		GitHubBaseURL:    viper.GetString("GITHUB_BASE_URL"),
		PostgresUser:     viper.GetString("POSTGRES_USER"),
		PostgresPassword: viper.GetString("POSTGRES_PASSWORD"),
		PostgresDB:       viper.GetString("POSTGRES_DB"),
		PostgresHost:     viper.GetString("POSTGRES_HOST"),
	}
}
