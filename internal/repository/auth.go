package repository

import (
	"api_chat_ws/dto"
	"api_chat_ws/model"

	"gorm.io/gorm"
)

type AuthRepo interface {
	Register(req *dto.RegisterReq) error
	LoginEmail(email string) (*model.User, error)
}

type authRepo struct {
	db *gorm.DB
}

func NewAuthRepo(db *gorm.DB) AuthRepo {
	return &authRepo{db}
}

func (r *authRepo) Register(req *dto.RegisterReq) error {
	newUser := model.User{
		Email:    req.Email,
		Password: req.Password,
		Username: req.Name,
	}

	return r.db.Model(&model.User{}).Create(&newUser).Error
}

func (r *authRepo) LoginEmail(email string) (*model.User, error) {
	var user model.User
	if err := r.db.Model(&model.User{}).Select("id", "email", "password").Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
