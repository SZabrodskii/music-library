package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"music-library/gateway/models"
	"music-library/utils/config"
)

func InitDB() (*gorm.DB, error) {
	dsn := "host=" + config.GetEnv("DB_HOST") + " user=" + config.GetEnv("DB_USER") + " password=" + config.GetEnv("DB_PASSWORD") + " dbname=" + config.GetEnv("DB_NAME") + " port=" + config.GetEnv("DB_PORT") + " sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	db.AutoMigrate(&models.Song{}, &models.Verse{})
	return db, nil
}
