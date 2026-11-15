package servise

import (
	"errors"
	"strings"

	"example/mymodule/internal/model"
	"example/mymodule/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var ErrLoginTaken = errors.New("login is already taken")
var ErrInvalidCredentials = errors.New("invalid login or password")

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(req model.CreateUserRequest) (*model.User, error) {
	req.Username = strings.TrimSpace(req.Username)
	req.Login = strings.ToLower(strings.TrimSpace(req.Login))
	if req.Username == "" || req.Login == "" || len(req.Userpass) < 8 || len(req.Userpass) > 72 {
		return nil, errors.New("invalid registration data")
	}
	if _, err := s.repo.GetByLogin(req.Login); err == nil {
		return nil, ErrLoginTaken
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashedPassword, err := HashPass(req.Userpass)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username: req.Username,
		Login:    req.Login,
		Userpass: hashedPassword,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Authenticate(login, password string) (*model.User, error) {
	user, err := s.repo.GetByLogin(strings.ToLower(strings.TrimSpace(login)))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Userpass), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (s *UserService) GetByID(id uint) (*model.User, error) {
	return s.repo.GetByID(id)
}

func HashPass(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
