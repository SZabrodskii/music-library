package main

import (
	"github.com/SZabrodskii/music-library/gateway/handlers"
	_ "github.com/SZabrodskii/music-library/song-service/migrations"
	"github.com/SZabrodskii/music-library/utils/providers"
	"github.com/SZabrodskii/music-library/utils/services"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.Provide(
			providers.NewLoggerProviderConfig,
			providers.NewLogger,
			providers.UseLogger,
			providers.NewCacheProvider,
			services.NewSongServiceClientConfig,
			services.NewSongServiceClient,
			providers.NewRedisProviderConfig,
			providers.NewRedisProvider,
			handlers.NewSongHandler,
			handlers.NewRouter,
		),
		fx.Invoke(startServer),
	)
	app.Run()
}

func startServer(router *handlers.Router) {
	router.Start()
}
