package repository

import (
	"errors"

	"example/mymodule/internal/model"
	"gorm.io/gorm"
)

var ErrAdminRequired = errors.New("administrator access required")
var ErrSelfDeletion = errors.New("cannot delete the current administrator")
var ErrLastAdminDeletion = errors.New("cannot delete the last administrator")

type DeletionRepository struct {
	db *gorm.DB
}

func NewDeletionRepository(db *gorm.DB) *DeletionRepository {
	return &DeletionRepository{db: db}
}

func (r *DeletionRepository) DeleteMusic(musicID, userID uint) (*model.Music, error) {
	var deleted model.Music
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		query := tx.Where("id = ?", musicID)
		if !user.IsAdmin {
			query = query.Where("username = ?", userID)
		}
		if err := query.First(&deleted).Error; err != nil {
			return err
		}

		if err := removeMusicReferences(tx, map[int]struct{}{int(musicID): {}}); err != nil {
			return err
		}
		return tx.Delete(&deleted).Error
	})
	if err != nil {
		return nil, err
	}
	return &deleted, nil
}

func (r *DeletionRepository) DeletePlaylist(playlistID, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		query := tx.Where("id = ?", playlistID)
		if !user.IsAdmin {
			query = query.Where("id_user = ?", userID)
		}
		result := query.Delete(&model.Playlist{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func (r *DeletionRepository) DeleteUser(targetID, actorID uint) ([]string, error) {
	var uploadPaths []string
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var actor model.User
		if err := tx.First(&actor, actorID).Error; err != nil {
			return err
		}
		if !actor.IsAdmin {
			return ErrAdminRequired
		}
		if targetID == actorID {
			return ErrSelfDeletion
		}

		var user model.User
		if err := tx.First(&user, targetID).Error; err != nil {
			return err
		}
		if user.IsAdmin {
			var adminCount int64
			if err := tx.Model(&model.User{}).Where("is_admin = ?", true).Count(&adminCount).Error; err != nil {
				return err
			}
			if adminCount <= 1 {
				return ErrLastAdminDeletion
			}
		}

		var music []model.Music
		if err := tx.Where("username = ?", targetID).Find(&music).Error; err != nil {
			return err
		}
		ids := make(map[int]struct{}, len(music))
		uploadPaths = make([]string, 0, len(music))
		for _, track := range music {
			ids[int(track.ID)] = struct{}{}
			uploadPaths = append(uploadPaths, track.PathToMusic)
		}
		if err := removeMusicReferences(tx, ids); err != nil {
			return err
		}
		if err := tx.Where("username = ?", targetID).Delete(&model.Music{}).Error; err != nil {
			return err
		}
		for _, record := range []interface{}{&model.Playlist{}, &model.Like{}, &model.Repost{}} {
			if err := tx.Where("id_user = ?", targetID).Delete(record).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&user).Error
	})
	if err != nil {
		return nil, err
	}
	return uploadPaths, nil
}

func removeMusicReferences(tx *gorm.DB, musicIDs map[int]struct{}) error {
	if len(musicIDs) == 0 {
		return nil
	}

	var likes []model.Like
	if err := tx.Find(&likes).Error; err != nil {
		return err
	}
	for _, like := range likes {
		like.ID_music = withoutMusicIDs(like.ID_music, musicIDs)
		if len(like.ID_music) == 0 {
			if err := tx.Delete(&like).Error; err != nil {
				return err
			}
		} else if err := tx.Save(&like).Error; err != nil {
			return err
		}
	}

	var playlists []model.Playlist
	if err := tx.Find(&playlists).Error; err != nil {
		return err
	}
	for _, playlist := range playlists {
		playlist.ID_music = withoutMusicIDs(playlist.ID_music, musicIDs)
		if err := tx.Save(&playlist).Error; err != nil {
			return err
		}
	}

	var reposts []model.Repost
	if err := tx.Find(&reposts).Error; err != nil {
		return err
	}
	for _, repost := range reposts {
		repost.ID_music = withoutMusicIDs(repost.ID_music, musicIDs)
		if len(repost.ID_music) == 0 {
			if err := tx.Delete(&repost).Error; err != nil {
				return err
			}
		} else if err := tx.Save(&repost).Error; err != nil {
			return err
		}
	}
	return nil
}

func withoutMusicIDs(values []int, removed map[int]struct{}) []int {
	filtered := make([]int, 0, len(values))
	for _, value := range values {
		if _, ok := removed[value]; !ok {
			filtered = append(filtered, value)
		}
	}
	return filtered
}
