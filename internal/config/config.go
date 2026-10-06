package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env-default:"local" env-required:"true" env:"ENV"`
	HttpServer struct {
		Host        string        `yaml:"host" env-default:"localhost" env-required:"true" env:"HTTP_SERVER_HOST"`
		Port        int           `yaml:"port" env-default:"8080" env-required:"true" env:"HTTP_SERVER_PORT"`
		Timeout     time.Duration `yaml:"timeout" env-default:"30" env-required:"true" env:"HTTP_SERVER_TIMEOUT"`
		IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60" env-required:"true" env:"HTTP_SERVER_IDLE_TIMEOUT"`
	} `yaml:"http_server"`
}

func NewConfig() *Config {
	configPath := "./config/local.yaml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("Config file %s does not exist, using default config", configPath)
		log.Fatalf("Config file %s does not exist, using default config", configPath)
	}
	var config Config
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		fmt.Printf("Failed to read config file %s: %v", configPath, err)
		log.Fatalf("Failed to read config file %s: %v", configPath, err)
	}
	return &config
}
