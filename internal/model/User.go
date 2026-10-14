package model


type User struct{
	ID uint `json:"id" gorm:"primaryKey"`
	Username string `json:"Username"`
	Login string `json:"Login"`
	Userpass string `json:"Userpass"`
	Avatar string `json:"Avatar"`
	
}

type CreateUserRequest struct{
	Username string `json:"Username"`
	Login string `json:"Login"`
	Userpass string `json:"Userpass"`
}