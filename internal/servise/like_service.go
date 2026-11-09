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

func (s *LikeService) CreateLike(req model.NewLikeRequest) (*model.Like, error) {
	like := &model.Like{
		ID_user:  req.ID_user,
		ID_music: []int{req.ID_music},
	}

	if err := s.repo.Create(like); err != nil {
		return nil, err
	}

	return like, nil

}

func (s *LikeService) GetAllForUser(User_id uint)([]model.Like, error){
	return s.repo.GetAllForUser(User_id)
}