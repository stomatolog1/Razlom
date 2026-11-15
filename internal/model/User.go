package model

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Username string `json:"username"`
	Login    string `json:"login" gorm:"uniqueIndex"`
	Userpass string `json:"-" gorm:"not null"`
	Avatar   string `json:"avatar"`
	IsAdmin  bool   `json:"isAdmin" gorm:"not null;default:false"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=1,max=60"`
	Login    string `json:"login" binding:"required,min=3,max=64"`
	Userpass string `json:"password" binding:"required,min=8,max=72"`
}
