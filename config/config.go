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
	viper.SetConfigFile(".env") // Specify the config file to read

	// Allow reading from environment variables
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("SERVER_ADDRESS", ":8080")
	viper.SetDefault("POLL_INTERVAL", 3600) // 1 hour in seconds
	viper.SetDefault("MAX_RETRIES", 3)
	viper.SetDefault("INITIAL_BACKOFF", 2) // In seconds
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379")
	viper.SetDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	// Read in .env file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			panic(err)
		}
		log.Println("No .env file found, using environment variables")
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
