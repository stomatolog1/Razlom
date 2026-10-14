package model

type Music struct {
	ID uint `json:"id" gorm:"primaryKey"`
	Name string `json:"Name"`
	PathToMusic string `json:"PathToMusic"`
	Username int `json:"Username"` // вместо артиста песни пишется тот кто её добавил на платформу ;3
	Genre string `json:"Genre"`
	Time int `json:"Time"`
	Listing int `json:"Listing"`
}

type NewMusicRequest struct{
	Name string `json:"Name"`
	Username int `json:"Username"`
	Genre string `json:"Genre"`
}