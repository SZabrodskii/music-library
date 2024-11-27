package providers

import (
	"github.com/SZabrodskii/music-library/utils"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

type LoggerProviderConfig struct {
	environment string
}

func NewLoggerProviderConfig() *LoggerProviderConfig {
	return &LoggerProviderConfig{
		environment: utils.GetEnv("ENVIRONMENT", "development"),
	}
}

func NewLogger(NewLoggerProviderConfig *LoggerProviderConfig) (*zap.Logger, error) {
	if NewLoggerProviderConfig.environment == "production" {
		return zap.NewProduction()
	} else {
		return zap.NewDevelopment()
	}
}

func UseLogger() fx.Option {
	return fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
		return &fxevent.ZapLogger{Logger: log}
	})
}
