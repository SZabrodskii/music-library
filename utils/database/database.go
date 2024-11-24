package database

import (
	"github.com/SZabrodskii/music-library/utils/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresProviderConfig struct {
	Host     string
	User     string
	Password string
	DBName   string
	Port     string
	SSLMode  string
}

func NewPostgresProviderConfig() *PostgresProviderConfig {
	return &PostgresProviderConfig{
		Host:     config.GetEnv("DB_HOST"),
		User:     config.GetEnv("DB_USER"),
		Password: config.GetEnv("DB_PASSWORD"),
		DBName:   config.GetEnv("DB_NAME"),
		Port:     config.GetEnv("DB_PORT"),
		SSLMode:  "disable",
	}
}

func InitDB(config *PostgresProviderConfig) (*gorm.DB, error) {
	dsn := "host=" + config.Host + " user=" + config.User + " password=" + config.Password + " dbname=" + config.DBName + " port=" + config.Port + " sslmode=" + config.SSLMode
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
