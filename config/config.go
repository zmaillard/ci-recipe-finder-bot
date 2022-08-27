package config

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type Config struct {
	SearchService string `env:"SearchService"`
	SearchIndex   string `env:"SearchIndex"`
	SearchApiKey  string `env:"SearchAPIKey"`

	Username string `env:"DatabaseUsername"`
	Password string `env:"DatabasePassword"`
	Engine   string `env:"DatabaseEngine"`
	Host     string `env:"DatabaseHost"`
	Port     string `env:"DatabasePort"`
	DbName   string `env:"DatabaseDbName"`
	SslMode  string `env:"DatabaseSSLMode"`

	SearchUIBaseUrl string `env:"SearchUIBaseUrl"`

	PhoneNumber string `env:"PhoneNumber"`
	PublicUrl   string `env:"PublicUrl"`
	HelpPage    string `env:"HelpPage"`

	TwilioFromNumber string `env:"TwilioFromNumber"`
	TwilioToNumbers  string `env:"TwilioToNumbers"`
	TwilioUser       string `env:"TwilioUser"`
	TwilioApiKey     string `env:"TwilioApiKey"`
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

func (c Config) BuildDatabaseUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:5432/%s", c.Username, c.Password, c.Host, c.DbName)
}
func (c Config) GetTwilioToNumbers() []string {
	return strings.Split(c.TwilioToNumbers, "^")
}

func (c Config) BuildTwilioUrl() string {
	return fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages", c.TwilioUser)
}
