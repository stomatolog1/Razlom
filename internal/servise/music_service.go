package servise

import (
	"example/mymodule/internal/model"
	"example/mymodule/internal/repository"
)

type MusicService struct {
	repo *repository.MusicRepository

}

func NewMusicrService(repo *repository.MusicRepository) *MusicService{
	return &MusicService{repo: repo}
}

func (s *MusicService) CreateMusic(req model.NewMusicRequest)(*model.Music, error){
	music := &model.Music{
		Name: req.Name,
		Username: req.Username,
		Genre: req.Genre,
	}
	
	if err:= s.repo.Create(music); err != nil{
		return nil, err
	}

	return music, nil
}


func (s *MusicService) GetAllByUser(UserID uint)([]model.Music, error){
	return s.repo.GetAllByUser(UserID)
}


func (s *MusicService) GetAllByGenre(Genre string)([]model.Music, error){
	return s.repo.GetAllByGenre(Genre)
}
