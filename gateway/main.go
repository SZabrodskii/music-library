package main

import (
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"music-library/consumer"
	_ "music-library/docs"
	"music-library/gateway/config"
	"music-library/gateway/handlers"
)

var logger *zap.Logger

func init() {
	logger, _ = zap.NewProduction()
	defer logger.Sync()
	config.LoadConfig()
}

func main() {
	app := fx.New(
		fx.Provide(
			config.LoadConfig,
		),
		fx.Invoke(
			func() {
				r := gin.Default()

				r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

				api := r.Group("/api/v1")
				{
					api.GET("/songs", handlers.GetSongs)
					api.GET("/songs/:songId/text", handlers.GetSongText)
					api.DELETE("/songs/:songId", handlers.DeleteSong)
					api.PATCH("/songs/:songId", handlers.UpdateSong)
					api.POST("/songs", handlers.AddSong)
				}

				go consumer.StartConsumer()

				r.Run(":8080")
			},
		),
	)

	app.Run()
}
