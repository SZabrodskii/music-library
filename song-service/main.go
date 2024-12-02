package main

import (
	"github.com/SZabrodskii/music-library/song-service/handlers"
	"github.com/SZabrodskii/music-library/song-service/migrations"
	"github.com/SZabrodskii/music-library/song-service/services"
	"github.com/SZabrodskii/music-library/utils/providers"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"log"
)

func main() {

	app := fx.New(
		fx.Provide(
			providers.NewLoggerProviderConfig,
			providers.NewLogger,
			providers.UseLogger,
			providers.NewRedisProviderConfig,
			providers.NewRedisProvider,
			providers.NewCacheProvider,
			providers.NewRabbitMQProviderConfig,
			providers.NewRabbitMQProvider,
			providers.NewPostgresProviderConfig,
			providers.NewPostgresProvider,
			services.NewSongServiceConfig,
			services.NewSongService,
			providers.NewJaegerProviderConfig,
			providers.NewJaegerProvider,
		),
		providers.JaegerProviderModule(),
		fx.Invoke(applyMigrations, handlers.RegisterHandlers),
	)

	app.Run()
}

func applyMigrations(db *gorm.DB) {
	if err := migrations.ApplyMigrations(db); err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}
}
