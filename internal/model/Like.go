package model

type Like struct {
	ID       uint  `json:"id" gorm:"primaryKey"`
	ID_music []int `json:"id_music"`
	ID_user  int   `json:"id_user"`
	Time     int   `json:"Time"`
}

type NewLikeRequest struct {
	ID_user  int `json:"ID_user"`
	ID_music int `json:"ID_music"`
}
