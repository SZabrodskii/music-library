package services

import (
	"encoding/json"
	"fmt"
	"github.com/SZabrodskii/music-library/utils/models"
	"github.com/SZabrodskii/music-library/utils/providers"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type SongService struct {
	logger *zap.Logger
	db     *gorm.DB
	queue  *providers.RabbitMQProvider
}

func NewSongService(logger *zap.Logger, db *gorm.DB, queue *providers.RabbitMQProvider) *SongService {
	return &SongService{
		logger: logger,
		db:     db,
		queue:  queue,
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

func (s *SongService) ConsumeAddSongQueue() {
	s.queue.Consume("add_song_queue", func(d amqp.Delivery) {
		var song models.Song
		if err := json.Unmarshal(d.Body, &song); err != nil {
			s.logger.Error("Failed to unmarshal song", zap.Error(err))
			d.Nack(false, true)
			return
		}

		apiURL := os.Getenv("API_URL")
		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			s.logger.Error("Failed to create request", zap.Error(err))
			d.Nack(false, true)
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

		if resp.StatusCode != http.StatusOK {
			s.logger.Error("Failed to fetch song details", zap.String("status", resp.Status))
			d.Nack(false, true)
			return
		}

		var songDetail models.SongDetail
		if err := json.NewDecoder(resp.Body).Decode(&songDetail); err != nil {
			s.logger.Error("Failed to decode song details", zap.Error(err))
			d.Nack(false, true)
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

		verses := strings.Split(songDetail.Text, "\n\n")
		for _, verse := range verses {
			if err := tx.Create(&models.Verse{SongID: song.ID, Text: verse}).Error; err != nil {
				tx.Rollback()
				s.logger.Error("Failed to create verse", zap.Error(err))
				d.Nack(false, true)
				return
			}
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			s.logger.Error("Failed to commit transaction", zap.Error(err))
			d.Nack(false, true)
			return
		}

		d.Ack(false)
	})
}

func (s *SongService) ConsumeUpdateSongQueue() {
	s.queue.Consume("update_song_queue", func(d amqp.Delivery) {
		var req UpdateSongRequest
		if err := json.Unmarshal(d.Body, &req); err != nil {
			s.logger.Error("Failed to unmarshal update song request", zap.Error(err))
			d.Nack(false, true)
			return
		}

		if err := s.UpdateSong(&req); err != nil {
			s.logger.Error("Failed to update song", zap.Error(err))
			d.Nack(false, true)
			return
		}

		d.Ack(false)
	})
}

func (s *SongService) ConsumeDeleteSongQueue() {
	s.queue.Consume("delete_song_queue", func(d amqp.Delivery) {
		var req DeleteSongRequest
		if err := json.Unmarshal(d.Body, &req); err != nil {
			s.logger.Error("Failed to unmarshal delete song request", zap.Error(err))
			d.Nack(false, true)
			return
		}

		if err := s.DeleteSong(&req); err != nil {
			s.logger.Error("Failed to delete song", zap.Error(err))
			d.Nack(false, true)
			return
		}

		d.Ack(false)
	})
}

//create client that addresses song_service/ It should be in utils
