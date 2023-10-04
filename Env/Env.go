package Env

import (
	"os"

	"github.com/joho/godotenv"
)

var Reader, err = ReadDotEnv()

func ReadDotEnv() (ENV, error) {
	var err = godotenv.Load()
	if err != nil {
		return ENV{}, err
	}
	return ENV{
		DB_URL:   osGetenv("DB_URL", "mongodb://localhost:27017"),
		DB_NAME:  osGetenv("DB_NAME", "test"),
		APP_PORT: osGetenv("APP_PORT", "8080"),
	}, nil
}

type ENV struct {
	DB_URL   string
	DB_NAME  string
	APP_PORT string
}

func osGetenv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
