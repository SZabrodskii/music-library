package services

import (
	"encoding/json"
	"github.com/SZabrodskii/music-library/song-service/models"
	"github.com/SZabrodskii/music-library/utils/queue"
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
	conn   *amqp.Connection
}

func NewSongService(logger *zap.Logger, db *gorm.DB, queue *queue.Queue) *SongService {
	return &SongService{
		logger: logger,
		db:     db,
		conn:   queue.GetConnection(),
	}
}

func (s *SongService) GetSongs(page, pageSize string, filters []string) ([]models.Song, error) {
	var songs []models.Song
	pageInt, _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)
	query := s.db.Offset((pageInt - 1) * pageSizeInt).Limit(pageSizeInt)
	for _, filter := range filters {
		query = query.Where(filter)
	}
	if err := query.Find(&songs).Error; err != nil {
		return nil, err
	}

	return songs, nil
}

func (s *SongService) GetSongText(songId, page, pageSize string) ([]models.Verse, error) {
	var verses []models.Verse
	pageInt, _ := strconv.Atoi(page)
	pageSizeInt, _ := strconv.Atoi(pageSize)
	if err := s.db.Where("song_id = ?", songId).Offset((pageInt - 1) * pageSizeInt).Limit(pageSizeInt).Find(&verses).Error; err != nil {
		return nil, err
	}
	return verses, nil
}

func (s *SongService) DeleteSong(songId string) error {
	if err := s.db.Where("id = ?", songId).Delete(&models.Song{}).Error; err != nil {
		return err
	}

	return nil
}

func (s *SongService) UpdateSong(songId string, song *models.Song) error {
	if err := s.db.Where("id = ?", songId).Updates(song).Error; err != nil {
		return err
	}

	return nil
}

func (s *SongService) AddSongToQueue(song *models.Song) error {
	ch, err := s.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"song_queue", // name
		false,        // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(song)
	if err != nil {
		return err
	}

	if err := ch.Publish(
		"",     // exchange
		q.Name, //routing key
		false,  //mandatory
		false,  //immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		}); err != nil {
		return err
	}
	return nil
}

func (s *SongService) ConsumeSongQueue() {
	ch, err := s.conn.Channel()
	if err != nil {
		s.logger.Fatal("Failed to open the channel", zap.Error(err))
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"song_queue", // name
		false,        //durable
		false,        //delete when unused
		false,        //exclusive
		false,        //no-wait
		nil,          //arguments
	)

	if err != nil {
		s.logger.Fatal("Failed to declare a queue", zap.Error(err))
	}

	msgs, err := ch.Consume(
		q.Name, //queue
		"",     //consumer
		false,  //auto ack
		false,  //exclusive
		false,  //no local
		false,  //no wait
		nil,    //args
	)
	if err != nil {
		s.logger.Fatal("Failed to register a consumer", zap.Error(err))
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var song models.Song
			if err := json.Unmarshal(d.Body, &song); err != nil {
				s.logger.Error("Failed to unmarshal song", zap.Error(err))
				d.Nack(false, true)
				continue
			}

			apiURL := os.Getenv("API_URL")
			req, err := http.NewRequest("GET", apiURL, nil)
			if err != nil {
				s.logger.Error("Failed to create request", zap.Error(err))
				d.Nack(false, true)
				continue
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
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				s.logger.Error("Failed to fetch song details", zap.String("status", resp.Status))
				d.Nack(false, true)
				continue
			}

			var songDetail models.SongDetail
			if err := json.NewDecoder(resp.Body).Decode(&songDetail); err != nil {
				s.logger.Error("Failed to decode song details", zap.Error(err))
				d.Nack(false, true)
				continue
			}

			song.ReleaseDate = songDetail.ReleaseDate
			song.Link = songDetail.Link

			tx := s.db.Begin()
			if err := tx.Create(&song).Error; err != nil {
				tx.Rollback()
				s.logger.Error("Failed to create song", zap.Error(err))
				d.Nack(false, true)
				continue
			}

			verses := strings.Split(songDetail.Text, "\n\n")
			for _, verse := range verses {
				if err := tx.Create(&models.Verse{SongID: song.ID, Text: verse}).Error; err != nil {
					tx.Rollback()
					s.logger.Error("Failed to create verse", zap.Error(err))
					d.Nack(false, true)
					continue
				}
			}

			if err := tx.Commit().Error; err != nil {
				tx.Rollback()
				s.logger.Error("Failed to commit transaction", zap.Error(err))
				d.Nack(false, true)
				continue
			}

			d.Ack(false)
		}
	}()

	<-forever
}
