package main

import (
	"github.com/SZabrodskii/music-library/gateway/handlers"
	"github.com/SZabrodskii/music-library/song-service/services"
	"github.com/SZabrodskii/music-library/utils/cache"
	"github.com/SZabrodskii/music-library/utils/config"
	"github.com/SZabrodskii/music-library/utils/database"
	"github.com/SZabrodskii/music-library/utils/queue"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	_ "music-library/gateway/docs"
)

func main() {
	app := fx.New(
		fx.Provide(
			NewLogger,
			cache.NewCache,
			config.GetEnv,
			database.InitDB,
			queue.NewQueue,
			services.NewSongService,
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

func startServer(router *gin.Engine) {
	router.Run(":8080")
}
