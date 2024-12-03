package handlers

import (
	"context"
	"errors"
	internalServices "github.com/SZabrodskii/music-library/song-service/services"
	"github.com/SZabrodskii/music-library/utils/middleware"
	"github.com/SZabrodskii/music-library/utils/models"
	"github.com/SZabrodskii/music-library/utils/providers"
	"github.com/SZabrodskii/music-library/utils/services"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
	"os"
	"time"
)

type SongHandler struct {
	cache   *providers.CacheProvider
	service *internalServices.SongService
	logger  *zap.Logger
}

func NewSongHandler(cache *providers.CacheProvider, service *internalServices.SongService, logger *zap.Logger, tracer trace.Tracer) *SongHandler {
	return &SongHandler{
		cache:   cache,
		service: service,
		logger:  logger,
	}
}

// GetSongs godoc
// @Summary Get songs with filtering and pagination
// @Description Get songs with filtering and pagination
// @Tags songs
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Limit per page" default(10)
// @Param filters query []string false "Filters"
// @Success 200 {array} models.Song
// @Router /songs [get]
func (h *SongHandler) GetSongs(c *gin.Context) {

	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")
	filters := c.QueryArray("filters")
	h.logger.Debug("Got req to get songs",
		zap.String("page", page),
		zap.String("pageSize", pageSize),
		zap.Strings("filters", filters),
		zap.String("traceparent",
			c.Request.Header.Get("traceparent")))

	request := &internalServices.GetSongsRequest{
		Page:     page,
		PageSize: pageSize,
		Filters:  filters,
	}

	songs, err := h.service.GetSongs(request)
	if err != nil {
		h.logger.Error("Failed to get songs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.logger.Debug("Get songs request has ended successfully",
		zap.String("page", page),
		zap.String("pageSize", pageSize),
		zap.Strings("filters", filters),
		zap.String("traceparent", c.Request.Header.Get("traceparent")))
	c.JSON(http.StatusOK, services.GetSongsResponse{Songs: songs})
}

// GetSongText godoc
// @Summary Get song text with pagination by verses
// @Description Get song text with pagination by verses
// @Tags songs
// @Accept json
// @Produce json
// @Param songId path int true "Song ID"
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Limit per page" default(10)
// @Success 200 {array} models.Verse
// @Router /songs/{songId}/text [get]
func (h *SongHandler) GetSongText(c *gin.Context) {
	songId := c.Param("songId")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")

	h.logger.Debug("Got req to get song text",
		zap.String("songId", songId),
		zap.String("page", page),
		zap.String("pageSize", pageSize),
		zap.String("traceparent", c.Request.Header.Get("traceparent")))

	request := &internalServices.GetSongTextRequest{
		SongId:   songId,
		Page:     page,
		PageSize: pageSize,
	}

	verses, err := h.service.GetSongText(request)
	if err != nil {
		h.logger.Error("Failed to get song text", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug("Got song text req has ended",
		zap.String("songId", songId),
		zap.String("page", page),
		zap.String("pageSize", pageSize),
		zap.String("traceparent", c.Request.Header.Get("traceparent")))
	c.JSON(http.StatusOK, verses)
}

// DeleteSong godoc
// @Summary Delete a song
// @Description Delete a song by ID
// @Tags songs
// @Accept json
// @Produce json
// @Param songId path int true "Song ID"
// @Success 204
// @Router /songs/{songId} [delete]
func (h *SongHandler) DeleteSong(c *gin.Context) {
	songId := c.Param("songId")

	h.logger.Debug("Got req to delete song",
		zap.String("songId", songId),
		zap.String("traceparent", c.Request.Header.Get("traceparent")))

	request := &internalServices.DeleteSongRequest{
		SongId: songId,
	}

	if err := h.service.PublishToQueue("delete_song_queue", request); err != nil {
		h.logger.Error("Failed to publish delete song task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug("Delete song req has ended",
		zap.String("songId", songId),
		zap.String("traceparent", c.Request.Header.Get("traceparent")))

	h.cache.DeleteFromCache("song_" + songId)
	c.Status(http.StatusNoContent)
}

// UpdateSong godoc
// @Summary Update a song
// @Description Update a song by ID
// @Tags songs
// @Accept json
// @Produce json
// @Param songId path int true "Song ID"
// @Param song body models.Song true "Song data"
// @Success 200 {object} models.Song
// @Router /songs/{songId} [patch]
func (h *SongHandler) UpdateSong(c *gin.Context) {
	songId := c.Param("songId")
	var song models.Song
	if err := c.ShouldBindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug("Got req to update song",
		zap.String("songId", songId),
		zap.Any("song", song),
		zap.String("traceparent", c.Request.Header.Get("traceparent")))

	request := &internalServices.UpdateSongRequest{
		SongID: songId,
		Song:   &song,
	}

	if err := h.service.PublishToQueue("update_song_queue", request); err != nil {
		h.logger.Error("Failed to publish update song task", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug("Update song req has ended",
		zap.String("songId", songId),
		zap.Any("song", song),
		zap.String("traceparent", c.Request.Header.Get("traceparent")))

	h.cache.DeleteFromCache("song_" + songId)
	c.JSON(http.StatusOK, song)
}

// AddSong godoc
// @Summary Add a new song
// @Description Add a new song
// @Tags songs
// @Accept json
// @Produce json
// @Param song body models.Song true "Song data"
// @Success 204
// @Router /songs [post]
func (h *SongHandler) AddSong(c *gin.Context) {
	var song models.Song
	if err := c.ShouldBindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug("Got req to add song",
		zap.Any("song", song),
		zap.String("traceparent", c.Request.Header.Get("traceparent")))

	request := &internalServices.AddSongRequest{
		Song: &song,
	}

	if err := h.service.AddSongToQueue(request); err != nil {
		h.logger.Error("Failed to add song to queue", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug("Add song req has ended",
		zap.Any("song", song),
		zap.String("traceparent", c.Request.Header.Get("traceparent")))

	h.cache.ClearCache()
	c.Status(http.StatusNoContent)
}

func RegisterHandlers(
	logger *zap.Logger,
	cache *providers.CacheProvider,
	songService *internalServices.SongService,
	lifecycle fx.Lifecycle,
	tracer trace.Tracer,
) *gin.Engine {
	handler := NewSongHandler(cache, songService, logger, tracer)
	router := gin.New()
	router.Use(middleware.TraceParentMiddleware())
	router.Use(gin.Recovery())
	router.Use(func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		duration := time.Since(start)
		logger.Info("Request completed",
			zap.String("traceparent", ctx.Request.Header.Get("traceparent")),
			zap.String("method", ctx.Request.Method),
			zap.String("path", ctx.Request.URL.Path),
			zap.Any("query", ctx.Request.URL.Query()),
			zap.Duration("duration", duration),
		)
	})
	router.Use(middleware.CacheMiddleware(cache))

	router.GET("/songs", handler.GetSongs)
	router.GET("/songs/:songId/text", handler.GetSongText)
	router.DELETE("/songs/:songId", handler.DeleteSong)
	router.PATCH("/songs/:songId", handler.UpdateSong)
	router.POST("/songs", handler.AddSong)

	lifecycle.Append(fx.Hook{
		OnStart: func(context.Context) error {
			songService.RegisterConsumers()
			songService.ConsumerManager.StartConsumers()
			return nil
		},
		OnStop: func(context.Context) error {
			return nil
		},
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	go func() {
		if err := router.Run(":" + port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	return router
}
