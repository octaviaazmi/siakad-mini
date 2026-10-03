package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort         string
	AppEnv          string
	DBHost          string
	DBPort          string
	DBUser          string
	DBPass          string
	DBName          string
	DBSSL           string
	JWTSecret       string
	JWTExpiredHours int
}

func Load() *Config {
	_ = godotenv.Load()

	expHours, _ := strconv.Atoi(getEnv("JWT_EXPIRED_HOURS", "24"))

	return &Config{
		AppPort:         getEnv("APP_PORT", "8080"),
		AppEnv:          getEnv("APP_ENV", "development"),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "5432"),
		DBUser:          getEnv("DB_USER", "siakad"),
		DBPass:          getEnv("DB_PASSWORD", "siakad123"),
		DBName:          getEnv("DB_NAME", "siakad_mini"),
		DBSSL:           getEnv("DB_SSLMODE", "disable"),
		JWTSecret:       getEnv("JWT_SECRET", "secret"),
		JWTExpiredHours: expHours,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
