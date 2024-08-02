package main

import (
	"fmt"
	"os"
)

type Config struct {
	ListenAddress string
	Port          string
	DBUser        string
	DBPassword    string
	DBAddress     string
	DBName        string
	JWTSecret     string
}

var Envs = initConfig()

func initConfig() Config {
	return Config{
		ListenAddress: getEnv("LISTEN_ADDRESS", "127.0.0.1"),
		Port:          getEnv("PORT", "3000"),
		DBUser:        getEnv("DB_USER", "root"),
		DBPassword:    getEnv("DB_PASSWORD", "P@ssw0rd"),
		DBAddress:     fmt.Sprintf("%s:%s", getEnv("DB_HOST", "127.0.0.1"), getEnv("DB_PORT", "3306")),
		DBName:        getEnv("DB_NAME", "project-manager"),
		JWTSecret:     getEnv("JWT_SECRET", "2xFavbztyHyRVFxuWrwtPtSQuwuQ1Y9i"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
