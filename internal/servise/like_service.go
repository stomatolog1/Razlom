package servise

import (
	"example/mymodule/internal/model"
	"example/mymodule/internal/repository"
)

type LikeService struct {
	repo *repository.LikeRepository
}

func NewLikeService(repo *repository.LikeRepository) *LikeService {
	return &LikeService{repo: repo}
}

func (s *LikeService) CreateLike(userID uint, req model.NewLikeRequest) (*model.Like, error) {
	existingLikes, err := s.repo.GetAllForUser(userID)
	if err != nil {
		return nil, err
	}
	for _, existing := range existingLikes {
		for _, musicID := range existing.ID_music {
			if musicID == req.ID_music {
				return &existing, nil
			}
		}
	}

	like := &model.Like{
		ID_user:  int(userID),
		ID_music: []int{req.ID_music},
	}

	if err := s.repo.Create(like); err != nil {
		return nil, err
	}

	return like, nil

}

func (s *LikeService) GetAllForUser(User_id uint) ([]model.Like, error) {
	return s.repo.GetAllForUser(User_id)
}

func (s *LikeService) DeleteLike(userID uint, musicID int) error {
	return s.repo.DeleteForUser(userID, musicID)
}
