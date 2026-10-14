package servise

import (
	"example/mymodule/internal/model"
	"example/mymodule/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService{
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(req model.CreateUserRequest)(*model.User, error){
	user := &model.User{
		Username: req.Username,
		Userpass: HashPass(req.Userpass),
	}

	if err:= s.repo.Create(user); err != nil{
		return nil, err
	}

	return user, nil

}


func HashPass(p string) string{
	return p
}
