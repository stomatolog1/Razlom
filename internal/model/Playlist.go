package model

type playlist struct {
	ID uint `json:"id" gorm:"primaryKey"`
	ID_music []int `json:"id_music"`
	ID_user int `json:"id_user"`
	Name string `json:"Name"`
	Private bool `json:"Private"`
}

type NewPlaylistRequest struct{
	ID_user int `json:"id_user"`
	Name string `json:"Name"`
	Private bool `json:"Private"`
}