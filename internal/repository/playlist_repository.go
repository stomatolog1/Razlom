package repository

import (
	"example/mymodule/internal/model"

	"gorm.io/gorm"
)

type PlaylistRepository struct {
	db *gorm.DB
}

func NewPlaylistRepository(db *gorm.DB) *PlaylistRepository {
	return &PlaylistRepository{db: db}
}

func (r *PlaylistRepository) Create(playlist *model.Playlist) error {
	return r.db.Create(playlist).Error
}

func (r *PlaylistRepository) GetAllPublic() ([]model.Playlist, error) {
	playlists := make([]model.Playlist, 0)
	if err := r.db.Where("private = ?", false).Order("id DESC").Find(&playlists).Error; err != nil {
		return nil, err
	}
	return playlists, nil
}

func (r *PlaylistRepository) GetAll() ([]model.Playlist, error) {
	playlists := make([]model.Playlist, 0)
	if err := r.db.Order("id DESC").Find(&playlists).Error; err != nil {
		return nil, err
	}
	return playlists, nil
}

func (r *PlaylistRepository) GetPublicPlaylist(User_id uint) ([]model.Playlist, error) {
	playlist := make([]model.Playlist, 0)
	if err := r.db.Where("id_user = ? AND private = ?", User_id, false).Find(&playlist).Error; err != nil {
		return nil, err
	}
	return playlist, nil
}

func (r *PlaylistRepository) GetByUser(User_id uint) ([]model.Playlist, error) {
	playlist := make([]model.Playlist, 0)
	if err := r.db.Where("id_user = ?", User_id).Find(&playlist).Error; err != nil {
		return nil, err
	}
	return playlist, nil
}

func (r *PlaylistRepository) GetByName(User_id uint, Name string) ([]model.Playlist, error) {
	playlists := make([]model.Playlist, 0)

	if err := r.db.Where("name = ? AND (private = ? OR id_user = ?)", Name, false, User_id).
		Find(&playlists).Error; err != nil {
		return nil, err
	}

	return playlists, nil
}

func (r *PlaylistRepository) AddMusic(userID, playlistID uint, musicID int) (*model.Playlist, error) {
	var playlist model.Playlist
	if err := r.db.Where("id = ? AND id_user = ?", playlistID, userID).First(&playlist).Error; err != nil {
		return nil, err
	}
	for _, existingID := range playlist.ID_music {
		if existingID == musicID {
			return &playlist, nil
		}
	}
	playlist.ID_music = append(playlist.ID_music, musicID)
	if err := r.db.Save(&playlist).Error; err != nil {
		return nil, err
	}
	return &playlist, nil
}
