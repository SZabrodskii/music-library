package models

import "gorm.io/gorm"

type Verse struct {
	gorm.Model
	SongID uint   `json:"song_id"`
	Text   string `json:"text"`
}
