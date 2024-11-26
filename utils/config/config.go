package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

func init() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	log.Println("DB_HOST:", os.Getenv("DB_HOST"))
	log.Println("DB_PORT:", os.Getenv("DB_PORT"))
	log.Println("DB_USER:", os.Getenv("DB_USER"))
	log.Println("DB_PASSWORD:", os.Getenv("DB_PASSWORD"))
	log.Println("DB_NAME:", os.Getenv("DB_NAME"))
}

func GetEnv(key string) string {
	return os.Getenv(key)
}

func ProvideEnvConfig() map[string]string {
	return map[string]string{
		"DB_HOST":      GetEnv("DB_HOST"),
		"DB_USER":      GetEnv("DB_USER"),
		"DB_PASSWORD":  GetEnv("DB_PASSWORD"),
		"DB_NAME":      GetEnv("DB_NAME"),
		"DB_PORT":      GetEnv("DB_PORT"),
		"RABBITMQ_URL": GetEnv("RABBITMQ_URL"),
		"REDIS_URL":    GetEnv("REDIS_URL"),
		"API_URL":      GetEnv("API_URL"),
	}
}
