package model

type Playlist struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	ID_music []int  `json:"id_music" gorm:"serializer:json"`
	ID_user  int    `json:"id_user"`
	Name     string `json:"name"`
	Private  bool   `json:"private"`
}

type NewPlaylistRequest struct {
	Name    string `json:"name" binding:"required,max=60"`
	Private bool   `json:"private"`
}
