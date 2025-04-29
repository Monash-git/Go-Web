package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"context"

	"golang.org/x/crypto/bcrypt"
)

var ErrUserDuplicateEmail = repository.ErrUserDuplicateEmail

type UserService struct{
	repo *repository.UserRepository
}

//初始化UserService
func NewUserService(repo *repository.UserRepository) * UserService{
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) Signup(ctx context.Context,u domain.User)error{
	//加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password),bcrypt.DefaultCost)
	if err != nil{
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx,u)
}