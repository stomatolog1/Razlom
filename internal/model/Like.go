package model

type Like struct {
	ID       uint  `json:"id" gorm:"primaryKey"`
	ID_music []int `json:"id_music" gorm:"serializer:json"`
	ID_user  int   `json:"id_user"`
	Time     int   `json:"time"`
}

type NewLikeRequest struct {
	ID_music int `json:"id_music" binding:"required,min=1"`
}
