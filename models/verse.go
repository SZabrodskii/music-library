package models

type Verse struct {
	ID     uint   `json:"id" gorm:"primaryKey"`
	SongID uint   `json:"song_id"`
	Text   string `json:"text"`
}
