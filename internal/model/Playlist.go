package model

type Playlist struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	ID_music []int  `json:"id_music" gorm:"serializer:json"`
	ID_user  int    `json:"id_user"`
	Name     string `json:"name"`
	Private  bool   `json:"private"`
}

type NewPlaylistRequest struct {
	ID_user int    `json:"id_user" binding:"required,min=1"`
	Name    string `json:"name" binding:"required"`
	Private bool   `json:"private"`
}
