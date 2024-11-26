package main

import (
	"github.com/SZabrodskii/music-library/song-service/handlers"
	"github.com/SZabrodskii/music-library/song-service/migrations"
	"github.com/SZabrodskii/music-library/song-service/services"
	"github.com/SZabrodskii/music-library/utils/config"
	"github.com/SZabrodskii/music-library/utils/providers"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"log"
)

func main() {
	app := fx.New(
		fx.Provide(
			NewLogger,
			config.GetEnv,
			providers.NewRedisProviderConfig,
			providers.NewRedisProvider,
			providers.NewCacheProvider,
			providers.NewRabbitMQProviderConfig,
			providers.NewRabbitMQProvider,
			providers.NewPostgresProviderConfig,
			providers.NewPostgresProvider,
			services.NewSongServiceConfig,
			services.NewSongService,
			handlers.RegisterHandlers,
		),
		fx.Invoke(applyMigrations),
	)

	app.Run()
}

func NewLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}

func applyMigrations(db *gorm.DB) {
	if err := migrations.ApplyMigrations(db); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}
}
