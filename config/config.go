package config

import (
	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	SearchService           string `env:"SearchService"`
	SearchIndex             string `env:"SearchIndex"`
	SearchApiKey            string `env:"SearchAPIKey"`
}
var config *Config

func Init() {
	err := godotenv.Load()
	if err != nil {
		log.WithFields(log.Fields{
			"envFile": true,
		}).Warn("Error loading .env file, reading configuration from ENV")
	}

	config = &Config{}
	err = env.Parse(config)
	if err != nil {
		log.WithFields(log.Fields{
			"envFile": false,
		}).Fatal("Failed to parse ENV")
	}

}

func GetConfig() *Config {
	return config
}
