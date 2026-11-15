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

func (r *LikeRepository) DeleteForUser(userID uint, musicID int) error {
	likes, err := r.GetAllForUser(userID)
	if err != nil {
		return err
	}
	for _, like := range likes {
		updatedMusicIDs := make([]int, 0, len(like.ID_music))
		found := false
		for _, likedMusicID := range like.ID_music {
			if likedMusicID == musicID {
				found = true
				continue
			}
			updatedMusicIDs = append(updatedMusicIDs, likedMusicID)
		}
		if !found {
			continue
		}
		if len(updatedMusicIDs) == 0 {
			return r.db.Where("id = ? AND id_user = ?", like.ID, userID).Delete(&model.Like{}).Error
		}
		like.ID_music = updatedMusicIDs
		return r.db.Save(&like).Error
	}
	return gorm.ErrRecordNotFound
}
