package handlers

import (
	"github.com/SZabrodskii/music-library/song-service/models"
	"github.com/SZabrodskii/music-library/song-service/services"
	"github.com/SZabrodskii/music-library/utils/cache"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

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
// @Router /api/v1/songs [get]
func GetSongs(c *gin.Context, cache *cache.Cache, service *services.SongService, logger *zap.Logger) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")
	filters := c.QueryArray("filters")

	cacheKey := "songs_" + page + "_" + pageSize + "_" + strings.Join(filters, "_")
	if val, ok := cache.GetFromCache(cacheKey); ok {
		c.JSON(http.StatusOK, val)
		return
	}

	songs, err := service.GetSongs(page, pageSize, filters)
	if err != nil {
		logger.Error("Failed to get songs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.SetToCache(cacheKey, songs, time.Minute*5)
	c.JSON(http.StatusOK, songs)
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
// @Router /api/v1/songs/{songId}/text [get]
func GetSongText(c *gin.Context, cache *cache.Cache, service *services.SongService, logger *zap.Logger) {
	songId := c.Param("songId")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")

	cacheKey := "song_text_" + songId + "_" + page + "_" + pageSize
	if val, ok := cache.GetFromCache(cacheKey); ok {
		c.JSON(http.StatusOK, val)
		return
	}

	verses, err := service.GetSongText(songId, page, pageSize)
	if err != nil {
		logger.Error("Failed to get song text", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.SetToCache(cacheKey, verses, time.Minute*5)
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
// @Router /api/v1/songs/{songId} [delete]
func DeleteSong(c *gin.Context, cache *cache.Cache, service *services.SongService, logger *zap.Logger) {
	songId := c.Param("songId")

	if err := service.DeleteSong(songId); err != nil {
		logger.Error("Failed to delete song", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	cache.DeleteFromCache("song_" + songId)
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
// @Router /api/v1/songs/{songId} [patch]
func UpdateSong(c *gin.Context, cache *cache.Cache, service *services.SongService, logger *zap.Logger) {
	songId := c.Param("songId")
	var song models.Song
	if err := c.ShouldBindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.UpdateSong(songId, &song); err != nil {
		logger.Error("Failed to update song", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.DeleteFromCache("song_" + songId)
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
// @Router /api/v1/songs [post]
func AddSong(c *gin.Context, cache *cache.Cache, service *services.SongService, logger *zap.Logger) {
	var song models.Song
	if err := c.ShouldBindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := service.AddSongToQueue(&song); err != nil {
		logger.Error("Failed to add song to queue", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.ClearCache()
	c.Status(http.StatusNoContent)
}
