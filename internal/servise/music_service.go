package servise

import (
	"example/mymodule/internal/model"
	"example/mymodule/internal/repository"
)

type MusicService struct {
	repo *repository.MusicRepository
}

func NewMusicService(repo *repository.MusicRepository) *MusicService {
	return &MusicService{repo: repo}
}

func (s *MusicService) CreateMusic(req model.NewMusicRequest, userID uint, path string) (*model.Music, error) {
	music := &model.Music{
		Name:        req.Name,
		Artist:      req.Artist,
		Username:    int(userID),
		Genre:       req.Genre,
		PathToMusic: path,
	}

	if err := s.repo.Create(music); err != nil {
		return nil, err
	}

	return music, nil
}

func (s *MusicService) GetAll() ([]model.Music, error) {
	return s.repo.GetAll()
}

func (s *MusicService) GetAllByUser(UserID uint) ([]model.Music, error) {
	return s.repo.GetAllByUser(UserID)
}

func (s *MusicService) GetAllByGenre(Genre string) ([]model.Music, error) {
	return s.repo.GetAllByGenre(Genre)
}
