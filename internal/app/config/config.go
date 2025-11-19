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

func NewConfig() (*Config, error) {
	var err error

	configName := "config"
	// Загружаем переменные окружения из файла app.env
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

	cfg := &Config{}           // создаем объект конфига
	err = viper.Unmarshal(cfg) // читаем информацию из файла,
	// конвертируем и затем кладем в нашу переменную cfg
	if err != nil {
		return nil, err
	}

	// парсим JWT настройки
	expiresInStr := viper.GetString("JWT.ExpiresIn")
	expiresIn, err := time.ParseDuration(expiresInStr)
	if err != nil {
		return nil, err
	}
	cfg.JWT.ExpiresIn = expiresIn

	// парсим Redis настройки из .env
	cfg.Redis.Host = os.Getenv("REDIS_HOST")
	cfg.Redis.Port, err = strconv.Atoi(os.Getenv("REDIS_PORT"))
	if err != nil {
		return nil, fmt.Errorf("redis port must be int value: %w", err)
	}
	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")
	cfg.Redis.User = os.Getenv("REDIS_USER")

	// парсим Redis таймауты из config.toml
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

	log.Info("config parsed")

	return cfg, nil
}
