package repository

import (
	"example/mymodule/internal/model"

	"gorm.io/gorm"
)

type MusicRepository struct{
	db *gorm.DB
}

func NewMusicRepository(db *gorm.DB) *MusicRepository{
	return &MusicRepository{db: db}
}

func (r *MusicRepository) Create(music *model.Music) error{
	return r.db.Create(music).Error
}

func (r *MusicRepository) GetAllByUser(UserID uint) ([]model.Music, error){
	var music []model.Music
	if err:= r.db.Where("user_id = ?", UserID).Find(&music).Error; err != nil{
		return nil, err
	}

	return music, nil
}

func (r *MusicRepository) GetAllByGenre(Genre string) ([]model.Music, error){
	var music []model.Music
	if err:= r.db.Where("genre = ?", Genre).Find(&music).Error; err != nil{
		return nil, err
	}
	return  music, nil
}

