package model

type Repost struct {
	ID       uint  `json:"id" gorm:"primaryKey"`
	ID_user  int   `json:"id_user"`
	ID_music []int `json:"id_music" gorm:"serializer:json"`
}

type CreateRepostRequest struct {
	ID_music int `json:"id_music" binding:"required,min=1"`
}
