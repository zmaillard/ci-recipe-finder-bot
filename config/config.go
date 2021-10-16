package config

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"time"
)

type Config struct {
	SearchService           string `env:"SearchService"`
	SearchIndex             string `env:"SearchIndex"`
	SearchApiKey            string `env:"SearchAPIKey"`

	Username string `env:"DatabaseUsername"`
	Password string `env:"DatabasePassword"`
	Engine   string `env:"DatabaseEngine"`
	Host     string `env:"DatabaseHost"`
	Port     string `env:"DatabasePort"`
	DbName   string `env:"DatabaseDbName"`
	SslMode  string `env:"DatabaseSSLMode"`

	SearchUIBaseUrl string `env:"SearchUIBaseUrl"`
}
var config *Config
var db *sqlx.DB
var err error

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

	dbURI := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		config.Host, config.Port, config.Username, config.DbName, config.Password, config.SslMode)

	if db != nil {
		log.Info("Using Existing Database")
		return
	}
	log.WithFields(log.Fields{
		"uri": dbURI,
	}).Info("Connecting To Database")
	db, err = sqlx.Open("postgres", dbURI)
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)

	d, err := time.ParseDuration("10m")
	if err != nil {
		log.WithFields(log.Fields{
			"uri": dbURI,
		}).Fatal("Cannot generate duration")
	}

	db.SetConnMaxLifetime(d)
	if err != nil {
		log.WithFields(log.Fields{
			"uri": dbURI,
		}).Fatal("Could Not Connect To Database")
	}

}

func GetConfig() *Config {
	return config
}


func GetDB() *sqlx.DB {
	return db
}
