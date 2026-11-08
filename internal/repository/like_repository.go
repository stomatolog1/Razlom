package repository

import (
	"example/mymodule/internal/model"

	"gorm.io/gorm"
)

type LikeRepository struct{
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) *LikeRepository{
	return &LikeRepository{db: db}
}

func (r *LikeRepository) Create(like *model.Like) error{
	return r.db.Create(like).Error
}

func (r *LikeRepository) GetAllForUser(User_id uint) ([]model.Like, error){
	var like []model.Like
	if err:= r.db.Where("user_id = ?", User_id).Find(&like).Error; err != nil{
		return nil, err
	}

	return like, nil
} 