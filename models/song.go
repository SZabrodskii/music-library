package models

type Song struct {
	ID          uint   `json:"id" gorm:"primary_key"`
	GroupName   string `json:"group"`
	SongName    string `json:"song"`
	ReleaseDate string `json:"releaseDate"`
	Link        string `json:"link"`
}

type SongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}
