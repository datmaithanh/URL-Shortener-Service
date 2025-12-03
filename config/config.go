package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	DOMAIN_NAME         = "https://short.datmt.id.vn"
	SERVER_ADDRESS      = ":8080"
	DB_DRIVER           = "postgres"
	URL_EXPIRE_DURATION = 24 * 7 * time.Hour
)

var (
	DBSource = getDBSource()
)

var envLoaded = false

func init() {
	if !envLoaded {
		LoadConfig()
		envLoaded = true
	}
}

func getDBSource() string {
	if dbSource := os.Getenv("DB_SOURCE"); dbSource != "" {
		return dbSource
	}
	return ""
}

func LoadConfig() {
	godotenv.Load(".env.prod")
	DBSource = getDBSource()

}
