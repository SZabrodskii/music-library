package handlers

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"music-library/cache"
	"music-library/models"
	"music-library/services"
	"net/http"
	"strings"
	"time"
)

var logger *zap.Logger

func init() {
	logger, _ = zap.NewProduction()
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			logger.Error("Failed to sync logger", zap.Error(err))
		}
	}(logger)

}

// @Summary Get songs with filtering and pagination
// @Description Get songs with filtering and pagination
// @Tags songs
// @Accept  json
// @Produce  json
// @Param   page  query  int  false  "Page number"  default(1)
// @Param   pageSize  query  int  false  "Limit per page"  default(10)
// @Param   filters  query  []string  false  "Filters"
// @Success 200 {array} models.Song
// @Router /songs [get]

func GetSongs(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")
	filters := c.QueryArray("filters")

	cacheKey := "songs_" + page + "_" + pageSize + "_" + strings.Join(filters, "_")
	if val, ok := cache.GetFromCache(cacheKey); ok {
		c.JSON(http.StatusOK, val)
		return
	}

	songs, err := services.GetSongs(page, pageSize, filters)
	if err != nil {
		logger.Error("Failed to get songs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.SetToCache(cacheKey, songs, time.Minute*5)
	c.JSON(http.StatusOK, songs)
}

// @Summary Get song text with pagination by verses
// @Description Get song text with pagination by verses
// @Tags songs
// @Accept  json
// @Produce  json
// @Param   songId  path  int  true  "Song ID"
// @Param   page  query  int  false  "Page number"  default(1)
// @Param   pageSize  query  int  false  "Limit per page"  default(10)
// @Success 200 {array} models.Verse
// @Router /songs/{songId}/text [get]

func GetSongText(c *gin.Context) {
	songId := c.Param("songId")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")

	cacheKey := "song_text_" + songId + "_" + page + "_" + pageSize
	if val, ok := cache.GetFromCache(cacheKey); ok {
		c.JSON(http.StatusOK, val)
		return
	}

	verses, err := services.GetSongText(songId, page, pageSize)
	if err != nil {
		logger.Error("Failed to get song text", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.SetToCache(cacheKey, verses, time.Minute*5)
	c.JSON(http.StatusOK, verses)
}

// @Summary Delete a song
// @Description Delete a song by ID
// @Tags songs
// @Accept  json
// @Produce  json
// @Param   songId  path  int  true  "Song ID"
// @Success 204
// @Router /songs/{songId} [delete]

func DeleteSong(c *gin.Context) {
	songId := c.Param("songId")

	if err := services.DeleteSong(songId); err != nil {
		logger.Error("Failed to delete song", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	cache.DeleteFromCache("song_" + songId)
	c.Status(http.StatusNoContent)
}

// @Summary Update a song
// @Description Update a song by ID
// @Tags songs
// @Accept  json
// @Produce  json
// @Param   songId  path  int  true  "Song ID"
// @Param   song  body  models.Song  true  "Song data"
// @Success 200 {object} models.Song
// @Router /songs/{songId} [patch]

func UpdateSong(c *gin.Context) {
	songId := c.Param("songId")
	var song models.Song
	if err := c.ShouldBindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.UpdateSong(songId, &song); err != nil {
		logger.Error("Failed to update song", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.DeleteFromCache("song_" + songId)
	c.JSON(http.StatusOK, song)

}

// @Summary Add a new song
// @Description Add a new song
// @Tags songs
// @Accept  json
// @Produce  json
// @Param   song  body  models.Song  true  "Song data"
// @Success 204
// @Router /songs [post]

func AddSong(c *gin.Context) {
	var song models.Song
	if err := c.ShouldBindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.AddSongToQueue(&song); err != nil {
		logger.Error("Failed to add song to queue", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cache.ClearCache()
	c.Status(http.StatusNoContent)
}
