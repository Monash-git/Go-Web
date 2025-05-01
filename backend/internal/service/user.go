package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicateEmail = repository.ErrUserDuplicateEmail
	ErrInvalidUserOrPassword = errors.New("用户不存在或者密码错误")
)

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

func (svc *UserService) Login(ctx context.Context,email,password string)(domain.User,error){
	u, err := svc.repo.FindByEmail(ctx, email)
	//用户登录校验逻辑
	//是否存在该邮箱对应的用户
	if err == repository.ErrUserNotFound  {
		return domain.User{},ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{},err
	}
	//校验密码是否正确
	err = bcrypt.CompareHashAndPassword([]byte(u.Password),[]byte(password))
	if err != nil {
		return domain.User{},ErrInvalidUserOrPassword
	}
	return u,nil
}