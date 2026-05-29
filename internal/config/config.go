package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"log"
	"log/slog"
	"os"
	"time"
)

type Config struct {
	Env     string    `yaml:"env"`
	DB      PrimaryDB `yaml:"database"`
	RedisDB Redis     `yaml:"redis"`
	GRPC    GRPC      `yaml:"grpc"`
	Mail    Mail      `yaml:"mail"`
}

type Mail struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	SSL      bool   `yaml:"ssl"`
}

type PrimaryDB struct {
	Host        string `yaml:"host"`
	Port        string `yaml:"port"`
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	Database    string `yaml:"dbname"`
	SSL         string `yaml:"ssl"`
	MaxAttempts int    `yaml:"max_attempts"`
}

type Redis struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
	DBNumber int    `yaml:"db"`
	Retries  int    `yaml:"retries"`
}

type GRPC struct {
	Host    string        `yaml:"host"`
	Port    string        `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func InitConfig() *Config {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		envFile = ".env.dev"
	}
	fmt.Println("env name", envFile)
	if err := godotenv.Load(envFile); err != nil {
		if !os.IsNotExist(err) {
			slog.Error("ошибка при инициализации переменных окружения", slog.Any("err", err))
		}
	}
	configPath := os.Getenv("CONFIG_PATH")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("CONFIG_PATH does not exist:%s", configPath)
	}

	// Read YAML file and substitute ${VAR} with environment variables
	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	expanded := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	return &cfg
}
