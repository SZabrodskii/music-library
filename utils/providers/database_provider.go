package providers

import (
	"fmt"
	"github.com/SZabrodskii/music-library/utils"
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
		Host:     utils.GetEnv("DB_HOST", "postgres"),
		User:     utils.GetEnv("DB_USER", "admin"),
		Password: utils.GetEnv("DB_PASSWORD", "password"),
		DBName:   utils.GetEnv("DB_NAME", "library"),
		Port:     utils.GetEnv("DB_PORT", "5432"),
		SSLMode:  utils.GetEnv("DB_SSL_MODE", "disable"),
	}
}

func NewPostgresProvider(config *PostgresProviderConfig) (*gorm.DB, error) {
	dsn := "host=" + config.Host + " user=" + config.User + " password=" + config.Password + " dbname=" + config.DBName + " port=" + config.Port + " sslmode=" + config.SSLMode
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	res, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err = res.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	return db, nil
}
