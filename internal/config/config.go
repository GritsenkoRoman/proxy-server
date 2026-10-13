package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local" env-required:"true" env:"ENV"`
	StoragePath string `yaml:"connection_sting" env-required:"true"`
	WorkerCount int    `yaml:"worker_count" env-default:"10"`
	QueueSize   int    `yaml:"queue_size" env-default:"100"`
	HttpServer  struct {
		Host        string        `yaml:"host" env-default:"localhost" env-required:"true" env:"HTTP_SERVER_HOST"`
		Port        int           `yaml:"port" env-default:"8080" env-required:"true" env:"HTTP_SERVER_PORT"`
		Timeout     time.Duration `yaml:"timeout" env-default:"30" env-required:"true" env:"HTTP_SERVER_TIMEOUT"`
		IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60" env-required:"true" env:"HTTP_SERVER_IDLE_TIMEOUT"`
	} `yaml:"http_server"`
}

func NewConfig() (*Config, error) {
	configPath := "./config/local.yaml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, err
	}
	var config Config
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		return &Config{}, err
	}
	return &config, nil
}
