package servise

import (
	"example/mymodule/internal/model"
	"example/mymodule/internal/repository"
)

type RepostService struct {
	repo *repository.RepostRepository
}

func NewRepostService(repo *repository.RepostRepository) *RepostService {
	return &RepostService{repo: repo}
}

func (s *RepostService) CreateRepost(userID uint, req model.CreateRepostRequest) (*model.Repost, error) {
	existingReposts, err := s.repo.GetAllUserRepost(userID)
	if err != nil {
		return nil, err
	}
	for _, existing := range existingReposts {
		for _, musicID := range existing.ID_music {
			if musicID == req.ID_music {
				return &existing, nil
			}
		}
	}

	repost := &model.Repost{
		ID_user:  int(userID),
		ID_music: []int{req.ID_music},
	}

	if err := s.repo.Create(repost); err != nil {
		return nil, err
	}

	return repost, nil
}

func (s *RepostService) GetAllUserRepost(userID uint) ([]model.Repost, error) {
	return s.repo.GetAllUserRepost(userID)
}
