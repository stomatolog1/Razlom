package servise

import (
	"example/mymodule/internal/model"
	"example/mymodule/internal/repository"
)

type PlaylistService struct {
	repo *repository.PlaylistRepository
}

func NewPlaylistService(repo *repository.PlaylistRepository) *PlaylistService {
	return &PlaylistService{repo: repo}
}

func (s *PlaylistService) CreateNewPlaylist(req model.NewPlaylistRequest) (*model.Playlist, error) {
	playlist := &model.Playlist{
		ID_user: req.ID_user,
		Name: req.Name,
		Private: req.Private,
	}

	if err := s.repo.Create(playlist); err != nil {
		return nil, err
	}

	return playlist, nil

}

func (s *PlaylistService) GetPublicPlaylist(User_id uint) ([]model.Playlist, error){ //все доступные не приватные плейлисты
	return s.repo.GetPublicPlaylist(User_id)
}

func (s *PlaylistService) GetByUser(User_id uint) ([]model.Playlist, error){ // все плелисты одного человека в Профиле, даёт приватные плейлисты если запрос сделал хозяин этих плейлистов
	return s.repo.GetByUser(User_id)
}

func (s *PlaylistService) GetByName(User_id uint, Name string) ([]model.Playlist, error){ // все плелисты одного человека в Профиле, даёт приватные плейлисты если запрос сделал хозяин этих плейлистов
	return s.repo.GetByName(User_id, Name)
}


