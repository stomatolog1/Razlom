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

func (s *RepostService) CreateRepost(req model.CreateRepostRequest) (*model.Repost, error) {
	repost := &model.Repost{
		ID_user:  req.ID_user,
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
