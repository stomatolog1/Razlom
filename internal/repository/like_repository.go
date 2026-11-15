package repository

import (
	"example/mymodule/internal/model"

	"gorm.io/gorm"
)

type LikeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) *LikeRepository {
	return &LikeRepository{db: db}
}

func (r *LikeRepository) Create(like *model.Like) error {
	return r.db.Create(like).Error
}

func (r *LikeRepository) GetAllForUser(User_id uint) ([]model.Like, error) {
	like := make([]model.Like, 0)
	if err := r.db.Where("id_user = ?", User_id).Find(&like).Error; err != nil {
		return nil, err
	}

	return like, nil
}
