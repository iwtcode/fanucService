package fanucService

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
}

type AppConfig struct {
	Port    string
	GinMode string
	APIKey  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type KafkaConfig struct {
	Broker string
	Topic  string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		App: AppConfig{
			Port:    getEnv("APP_PORT", "8080"),
			GinMode: getEnv("GIN_MODE", "debug"),
			APIKey:  getEnv("API_KEY"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD"),
			Name:     getEnv("DB_NAME", "fanuc_db"),
		},
		Kafka: KafkaConfig{
			Broker: getEnv("KAFKA_BROKER"),
			Topic:  getEnv("KAFKA_TOPIC"),
		},
	}
}

func getEnv(key string, fallback ...string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	if len(fallback) > 0 {
		return fallback[0]
	}

	return ""
}
