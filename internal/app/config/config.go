package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	ServiceHost string
	ServicePort int
	MinIO       MinIOConfig
	JWT         JWTConfig
	Redis       RedisConfig
	SMTP        SMTPConfig
}

type MinIOConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

type JWTConfig struct {
	Token         string
	ExpiresIn     time.Duration
	SigningMethod string
}

type RedisConfig struct {
	Host        string
	Password    string
	Port        int
	User        string
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

func NewConfig() (*Config, error) {
	var err error

	configName := "config"
	_ = godotenv.Load("app.env")
	_ = godotenv.Load()
	if os.Getenv("CONFIG_NAME") != "" {
		configName = os.Getenv("CONFIG_NAME")
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("toml")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")
	viper.WatchConfig()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	expiresInStr := viper.GetString("JWT.ExpiresIn")
	expiresIn, err := time.ParseDuration(expiresInStr)
	if err != nil {
		return nil, err
	}
	cfg.JWT.ExpiresIn = expiresIn

	cfg.Redis.Host = os.Getenv("REDIS_HOST")
	cfg.Redis.Port, err = strconv.Atoi(os.Getenv("REDIS_PORT"))
	if err != nil {
		return nil, fmt.Errorf("redis port must be int value: %w", err)
	}
	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")
	cfg.Redis.User = os.Getenv("REDIS_USER")

	dialTimeoutStr := viper.GetString("Redis.DialTimeout")
	dialTimeout, err := time.ParseDuration(dialTimeoutStr)
	if err != nil {
		return nil, err
	}
	cfg.Redis.DialTimeout = dialTimeout

	readTimeoutStr := viper.GetString("Redis.ReadTimeout")
	readTimeout, err := time.ParseDuration(readTimeoutStr)
	if err != nil {
		return nil, err
	}
	cfg.Redis.ReadTimeout = readTimeout

	// SMTP config from environment with defaults
	cfg.SMTP.Host = getEnv("SMTP_HOST", "smtp.gmail.com")
	cfg.SMTP.Port = getEnvAsInt("SMTP_PORT", 587)
	cfg.SMTP.Username = getEnv("SMTP_USERNAME", "ebookstore1504@gmail.com")
	cfg.SMTP.Password = getEnv("SMTP_PASSWORD", "nlwq ujqh xzlz saie")
	cfg.SMTP.From = getEnv("SMTP_FROM", "ebookstore1504@gmail.com")

	log.Info("config parsed")

	return cfg, nil
}

// getEnv получает переменную окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает переменную окружения как int или возвращает значение по умолчанию
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
