package servise

import (
	"example/mymodule/internal/model"
	"example/mymodule/internal/repository"
)

type AdminService struct {
	users     *repository.UserRepository
	music     *repository.MusicRepository
	playlists *repository.PlaylistRepository
	deletions *repository.DeletionRepository
}

func NewAdminService(
	users *repository.UserRepository,
	music *repository.MusicRepository,
	playlists *repository.PlaylistRepository,
	deletions *repository.DeletionRepository,
) *AdminService {
	return &AdminService{users: users, music: music, playlists: playlists, deletions: deletions}
}

func (s *AdminService) GetUsers() ([]model.User, error) {
	return s.users.GetAll()
}

func (s *AdminService) GetMusic() ([]model.Music, error) {
	return s.music.GetAll()
}

func (s *AdminService) GetPlaylists() ([]model.Playlist, error) {
	return s.playlists.GetAll()
}

func (s *AdminService) DeleteUser(targetID, actorID uint) ([]string, error) {
	return s.deletions.DeleteUser(targetID, actorID)
}
