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
	Name     string `json:"name" binding:"required"`
	Artist   string `json:"artist" binding:"required"`
	Username int    `json:"username" binding:"required,min=1"`
	Genre    string `json:"genre" binding:"required"`
	Time     int    `json:"time"`
}
