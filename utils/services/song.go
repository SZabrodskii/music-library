package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SZabrodskii/music-library/utils"
	"github.com/SZabrodskii/music-library/utils/models"
	"net/http"
)

type GetSongsRequest struct {
	Page     string   `json:"page"`
	PageSize string   `json:"pageSize"`
	Filters  []string `json:"filters"`
}

type GetSongsResponse struct {
	Songs []*models.Song `json:"songs"`
}

type GetSongTextRequest struct {
	SongId   string `json:"songId"`
	Page     string `json:"page"`
	PageSize string `json:"pageSize"`
}

type GetSongTextResponse struct {
	Verses []models.Verse `json:"verses"`
}

type DeleteSongRequest struct {
	SongId string `json:"songId"`
}

type UpdateSongRequest struct {
	SongID string      `json:"songId"`
	Song   models.Song `json:"song"`
}

type AddSongRequest struct {
	Song models.Song `json:"song"`
}

type SongServiceClientConfig struct {
	baseURL string
}

func NewSongServiceClientConfig() *SongServiceClientConfig {
	return &SongServiceClientConfig{
		baseURL: utils.GetEnv("SONG_SERVICE_URL", "http://song-service:8080"),
	}

}

type SongServiceClient struct {
	BaseURL    string
	httpClient *http.Client
}

func NewSongServiceClient(config *SongServiceClientConfig) *SongServiceClient {
	return &SongServiceClient{
		BaseURL:    config.baseURL,
		httpClient: &http.Client{},
	}
}

func (c *SongServiceClient) GetSongs(req *GetSongsRequest) (*GetSongsResponse, error) {
	url := fmt.Sprintf("%s/songs?page=%s&pageSize=%s", c.BaseURL, req.Page, req.PageSize)
	for _, filter := range req.Filters {
		url += "&filters=" + filter
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get songs: %s", resp.Status)
	}

	var response GetSongsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *SongServiceClient) GetSongText(req *GetSongTextRequest) (*GetSongTextResponse, error) {
	url := fmt.Sprintf("%s/songs/%s/text?page=%s&pageSize=%s", c.BaseURL, req.SongId, req.Page, req.PageSize)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get song text: %s", resp.Status)
	}

	var response GetSongTextResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *SongServiceClient) UpdateSong(req *UpdateSongRequest) error {
	url := fmt.Sprintf("%s/songs/%s", c.BaseURL, req.SongID)
	body, err := json.Marshal(req.Song)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("PATCH", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update song: %s", resp.Status)
	}
	return nil
}

func (c *SongServiceClient) AddSong(req *AddSongRequest) error {
	url := fmt.Sprintf("%s/songs", c.BaseURL)
	body, err := json.Marshal(req.Song)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to add song: %s", resp.Status)
	}
	return nil
}

func (c *SongServiceClient) DeleteSong(req *DeleteSongRequest) error {
	url := fmt.Sprintf("%s/songs/%s", c.BaseURL, req.SongId)

	httpReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete song: %s", resp.Status)
	}
	return nil
}
