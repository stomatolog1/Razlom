package repository

import (
	"example/mymodule/internal/model"

	"gorm.io/gorm"
)

type RepostRepository struct{
	db *gorm.DB
}

func NewRepostRepository(db *gorm.DB) *RepostRepository{
	return &RepostRepository{db: db}
}

func (r *RepostRepository) Create(repost *model.Repost) error{
	return r.db.Create(repost).Error
}

func (r *RepostRepository) GetAllUserRepost(User_id uint) ([]model.Repost, error){
	var repost []model.Repost
	if err := r.db.Where("id_user = ? ", User_id).Find(&repost).Error; err != nil{
		return nil, err
	}
	return repost, nil
}