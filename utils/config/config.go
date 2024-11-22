package config

import (
	"github.com/joho/godotenv"
	"os"
)

func init() {
	_ = godotenv.Load()
}

func GetEnv(key string) string {
	return os.Getenv(key)
}
