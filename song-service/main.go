package main

import (
	"github.com/SZabrodskii/music-library/song-service/handlers"
	"github.com/SZabrodskii/music-library/song-service/migrations"
	"github.com/SZabrodskii/music-library/song-service/services"
	"github.com/SZabrodskii/music-library/utils/cache"
	"github.com/SZabrodskii/music-library/utils/config"
	"github.com/SZabrodskii/music-library/utils/database"
	"github.com/SZabrodskii/music-library/utils/queue"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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
			handlers.RegisterHandlers,
		),
		fx.Invoke(applyMigrations, startReverseProxy),
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

func startReverseProxy(logger *zap.Logger) {
	target, err := url.Parse("http://song-service:8081")
	if err != nil {
		logger.Fatal("Failed to parse target URL", zap.Error(err))
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	http.Handle("/", proxy)
	logger.Info("Starting reverse proxy")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Fatal("Failed to start reverse proxy", zap.Error(err))
	}
}
