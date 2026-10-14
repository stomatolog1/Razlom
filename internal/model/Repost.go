package model

type Repost struct{
	ID uint `json:"id" gorm:"primaryKey"`
	ID_user int `json:"ID_user"`
	ID_music []int `json:"ID_music"`	
}

type CreateRepostRequest struct{
	ID_user int `json:"ID_user"`
	ID_music int `json:"ID_music"`	
}