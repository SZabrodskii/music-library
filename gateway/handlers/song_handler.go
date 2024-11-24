package handlers

import (
	"github.com/SZabrodskii/music-library/utils/cache"
	"github.com/SZabrodskii/music-library/utils/models"
	"github.com/SZabrodskii/music-library/utils/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

type SongHandler struct {
	cache  *cache.Cache
	client *services.Client
	logger *zap.Logger
}

func NewSongHandler(cache *cache.Cache, client *services.Client, logger *zap.Logger) *SongHandler {
	return &SongHandler{
		cache:  cache,
		client: client,
		logger: logger,
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
// @Router /api/v1/songs [get]
func (h *SongHandler) GetSongs(c *gin.Context) {
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")
	filters := c.QueryArray("filters")

	request := &services.GetSongsRequest{
		Page:     page,
		PageSize: pageSize,
		Filters:  filters,
	}

	response, err := h.client.GetSongs(request)
	if err != nil {
		h.logger.Error("Failed to get songs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Set("responseData", response.Songs)
	c.JSON(http.StatusOK, response.Songs)
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
func (h *SongHandler) GetSongText(c *gin.Context) {
	songId := c.Param("songId")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")

	request := &services.GetSongTextRequest{
		SongId:   songId,
		Page:     page,
		PageSize: pageSize,
	}

	response, err := h.client.GetSongText(request)
	if err != nil {
		h.logger.Error("Failed to get song text", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Set("responseData", response.Verses)
	c.JSON(http.StatusOK, response.Verses)
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
func (h *SongHandler) DeleteSong(c *gin.Context) {
	songId := c.Param("songId")
	request := &services.DeleteSongRequest{
		SongId: songId,
	}

	if err := h.client.DeleteSong(request); err != nil {
		h.logger.Error("Failed to delete song", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
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
// @Router /api/v1/songs/{songId} [patch]
func (h *SongHandler) UpdateSong(c *gin.Context) {
	songId := c.Param("songId")
	var song models.Song
	if err := c.ShouldBindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request := &services.UpdateSongRequest{
		SongID: songId,
		Song:   song,
	}

	if err := h.client.UpdateSong(request); err != nil {
		h.logger.Error("Failed to update song", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.cache.DeleteFromCache("song_" + songId)
	c.Set("responseData", song)
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
func (h *SongHandler) AddSong(c *gin.Context) {
	var song models.Song
	if err := c.ShouldBindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	request := &services.AddSongRequest{
		Song: song,
	}

	if err := h.client.AddSong(request); err != nil {
		h.logger.Error("Failed to add song to queue", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.cache.ClearCache()
	c.Status(http.StatusNoContent)
}
