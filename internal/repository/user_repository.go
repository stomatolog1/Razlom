package repository

import (
	"example/mymodule/internal/model"

	"gorm.io/gorm"
)

type UserRepository struct{
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository{
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error{
	return r.db.Create(user).Error
}

