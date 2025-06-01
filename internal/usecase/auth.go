package usecase

import (
	"api_chat_ws/dto"
	"api_chat_ws/helper/utils"
	"api_chat_ws/internal/repository"
	"errors"
)

type AuthUsecase interface {
	Register(req *dto.RegisterReq) error
	Login(req *dto.LoginReq) (string, error)
}

type authUsecase struct {
	authRepo repository.AuthRepo
}

func NewAuthUsecase(authRepo repository.AuthRepo) AuthUsecase {
	return &authUsecase{authRepo}
}

func (u *authUsecase) Register(req *dto.RegisterReq) error {
	valid := utils.IsValidEmail(req.Email)
	if !valid {
		return utils.ErrInvalidEmail
	}
	hashsed, err := utils.HashPasswrd(req.Password)
	if err != nil {
		return err
	}

	req.Password = hashsed
	if err := u.authRepo.Register(req); err != nil {
		return err
	}
	return nil
}

func (u *authUsecase) Login(req *dto.LoginReq) (string, error) {
	valid := utils.IsValidEmail(req.Email)
	if !valid {
		return "", utils.ErrInvalidEmail
	}
	user, err := u.authRepo.LoginEmail(req.Email)
	if err != nil {
		return "", err
	}

	if valid := utils.ComparePassword(user.Password, req.Password); !valid {
		return "", errors.New("email dan password tidak cocok")
	}

	jwt, err := utils.GenerateJWT(user.Email, user.ID)
	if err != nil {
		return "", err
	}

	return jwt, nil
}
