package database

import (
	"github.com/SZabrodskii/music-library/song-service/models"
	"github.com/SZabrodskii/music-library/utils/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
