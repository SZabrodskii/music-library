package main

import (
	"github.com/SZabrodskii/music-library/gateway/handlers"
	"github.com/SZabrodskii/music-library/utils/cache"
	"github.com/SZabrodskii/music-library/utils/config"
	"github.com/SZabrodskii/music-library/utils/database"
	"github.com/SZabrodskii/music-library/utils/queue"
	"github.com/SZabrodskii/music-library/utils/services"
	"go.uber.org/fx"
	"go.uber.org/zap"

	_ "github.com/SZabrodskii/music-library/song-service/migrations"
)

func main() {
	app := fx.New(
		fx.Provide(
			NewLogger,
			cache.NewCache,
			config.GetEnv,
			database.InitDB,
			queue.NewQueue,
			services.NewSongServiceClient,
			handlers.NewSongHandler,
			handlers.NewRouter,
		),
		fx.Invoke(startServer),
	)

	app.Run()
}

func NewLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}

func startServer(router *handlers.Router) {
	router.Start()
}
