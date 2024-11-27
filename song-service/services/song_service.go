package services

import (
	"encoding/json"
	"fmt"
	"github.com/SZabrodskii/music-library/utils"
	"github.com/SZabrodskii/music-library/utils/models"
	"github.com/SZabrodskii/music-library/utils/providers"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
)

type SongServiceConfig struct {
	SongInfoAPIHost string
}

func NewSongServiceConfig() *SongServiceConfig {
	return &SongServiceConfig{
		SongInfoAPIHost: utils.GetEnv("SONG_INFO_API_HOST", "localhost:8081"),
	}
}

type SongService struct {
	logger          *zap.Logger
	db              *gorm.DB
	queue           *providers.RabbitMQProvider
	ConsumerManager *ConsumerManager
	config          *SongServiceConfig
}

type ConsumerManager struct {
	queue    *providers.RabbitMQProvider
	logger   *zap.Logger
	db       *gorm.DB
	handlers map[string]func(amqp.Delivery)
}

func NewSongService(logger *zap.Logger, db *gorm.DB, queue *providers.RabbitMQProvider, config *SongServiceConfig) *SongService {
	consumerManager := NewConsumerManager(logger, db, queue)
	return &SongService{
		logger:          logger,
		db:              db,
		queue:           queue,
		ConsumerManager: consumerManager,
		config:          config,
	}
}

func NewConsumerManager(logger *zap.Logger, db *gorm.DB, queue *providers.RabbitMQProvider) *ConsumerManager {
	return &ConsumerManager{
		queue:    queue,
		logger:   logger,
		db:       db,
		handlers: make(map[string]func(amqp.Delivery)),
	}
}

func (cm *ConsumerManager) RegisterHandler(queueName string, handler func(amqp.Delivery)) {
	cm.handlers[queueName] = handler
	cm.queue.Consume(queueName, handler)
}

func (cm *ConsumerManager) StartConsumers() {
	for queueName, handler := range cm.handlers {
		cm.queue.Consume(queueName, handler)
	}
}

type GetSongsRequest struct {
	Page     string   `json:"page"`
	PageSize string   `json:"pageSize"`
	Filters  []string `json:"filters"`
}

func (s *SongService) GetSongs(req *GetSongsRequest) ([]*models.Song, error) {
	songs := make([]*models.Song, 0)
	pageInt, _ := strconv.Atoi(req.Page)
	pageSizeInt, _ := strconv.Atoi(req.PageSize)

	filters := req.Filters
	query := s.db.Offset((pageInt - 1) * pageSizeInt).Limit(pageSizeInt)
	for _, filter := range filters {
		query = query.Where(filter)
	}
	if err := query.Find(&songs).Error; err != nil {
		return nil, err
	}

	return songs, nil
}

type GetSongTextRequest struct {
	SongId   string `json:"songId"`
	Page     string `json:"page"`
	PageSize string `json:"pageSize"`
}

func (s *SongService) GetSongText(req *GetSongTextRequest) ([]*models.Verse, error) {
	var verses []*models.Verse
	pageInt, _ := strconv.Atoi(req.Page)
	pageSizeInt, _ := strconv.Atoi(req.PageSize)
	if err := s.db.Where("song_id = ?", req.SongId).Offset((pageInt - 1) * pageSizeInt).Limit(pageSizeInt).Find(&verses).Error; err != nil {
		return nil, err
	}
	return verses, nil
}

type DeleteSongRequest struct {
	SongId string `json:"songId"`
}

func (s *SongService) DeleteSong(req *DeleteSongRequest) error {
	if err := s.db.Where("id = ?", req.SongId).Delete(&models.Song{}).Error; err != nil {
		return err
	}
	return nil
}

type UpdateSongRequest struct {
	SongID string       `json:"songId"`
	Song   *models.Song `json:"song"`
}

func (s *SongService) UpdateSong(req *UpdateSongRequest) error {
	if err := s.db.Where("id = ?", req.SongID).First(&models.Song{}).Error; err != nil {
		return fmt.Errorf("song not found: %w", err)
	}

	if err := s.db.Model(&models.Song{}).Where("id = ?", req.SongID).Updates(req.Song).Error; err != nil {
		return fmt.Errorf("failed to update song: %w", err)
	}

	return nil
}

type AddSongRequest struct {
	Song *models.Song `json:"song"`
}

func (s *SongService) AddSongToQueue(req *AddSongRequest) error {
	body, err := json.Marshal(req.Song)
	if err != nil {
		return err
	}
	return s.queue.Publish("add_song_queue", body)
}

func (s *SongService) PublishToQueue(queueName string, data interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return s.queue.Publish(queueName, body)
}

func (s *SongService) handleAddSong(d amqp.Delivery) {
	var song models.Song
	if err := json.Unmarshal(d.Body, &song); err != nil {
		s.logger.Error("Failed to unmarshal song", zap.Error(err))
		d.Reject(false)
		return
	}

	apiURL := s.config.SongInfoAPIHost + "/info"
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		s.logger.Error("Failed to create request", zap.Error(err))
		d.Reject(false)
		return
	}

	q := req.URL.Query()
	q.Add("group", song.GroupName)
	q.Add("song", song.SongName)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		s.logger.Error("Failed to fetch song details", zap.Error(err))
		d.Nack(false, true)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		s.logger.Error("Failed to fetch song details", zap.String("status", resp.Status))
		d.Nack(false, true)
		return
	} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		s.logger.Error("Failed to fetch song details", zap.String("status", resp.Status))
	}

	var songDetail models.SongDetail
	if err := json.NewDecoder(resp.Body).Decode(&songDetail); err != nil {
		s.logger.Error("Failed to decode song details", zap.Error(err))
		d.Reject(false)
		return
	}

	song.ReleaseDate = songDetail.ReleaseDate
	song.Link = songDetail.Link

	tx := s.db.Begin()
	if err := tx.Create(&song).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create song", zap.Error(err))
		d.Nack(false, true)
		return
	}

	var verses []*models.Verse

	for _, verse := range strings.Split(songDetail.Text, "\n\n") {
		verses = append(verses, &models.Verse{SongID: song.ID, Text: verse})
	}

	err = tx.Create(&verses).Error
	if err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create verses", zap.Error(err))
		d.Nack(false, true)
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to commit transaction", zap.Error(err))
		d.Nack(false, true)
		return
	}

	d.Ack(false)
}

func (s *SongService) handleUpdateSong(d amqp.Delivery) {
	var req UpdateSongRequest
	if err := json.Unmarshal(d.Body, &req); err != nil {
		s.logger.Error("Failed to unmarshal update song request", zap.Error(err))
		d.Reject(false)
		return
	}

	if err := s.UpdateSong(&req); err != nil {
		s.logger.Error("Failed to update song", zap.Error(err))
		d.Nack(false, true)
		return
	}

	d.Ack(false)
}

func (s *SongService) handleDeleteSong(d amqp.Delivery) {
	var req DeleteSongRequest
	if err := json.Unmarshal(d.Body, &req); err != nil {
		s.logger.Error("Failed to unmarshal delete song request", zap.Error(err))
		d.Reject(false)
		return
	}

	if err := s.DeleteSong(&req); err != nil {
		s.logger.Error("Failed to delete song", zap.Error(err))
		d.Nack(false, true)
		return
	}

	d.Ack(false)
}

func (s *SongService) RegisterConsumers() {
	s.ConsumerManager.RegisterHandler("add_song_queue", s.handleAddSong)
	s.ConsumerManager.RegisterHandler("update_song_queue", s.handleUpdateSong)
	s.ConsumerManager.RegisterHandler("delete_song_queue", s.handleDeleteSong)
}
