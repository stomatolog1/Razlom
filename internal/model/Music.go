package model

type Music struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name"`
	Artist      string `json:"artist"`
	PathToMusic string `json:"pathToMusic"`
	Username    int    `json:"username"`
	Genre       string `json:"genre"`
	Time        int    `json:"time"`
	Listing     int    `json:"listing"`
}

type NewMusicRequest struct {
	Name   string `form:"name" binding:"required,max=80"`
	Artist string `form:"artist" binding:"required,max=80"`
	Genre  string `form:"genre" binding:"required,max=40"`
}
