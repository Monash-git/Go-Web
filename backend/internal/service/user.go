package service

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserDuplicateEmail = repository.ErrDuplicateUser
	ErrInvalidUserOrPassword = errors.New("用户不存在或者密码错误")
)

type UserService interface{
	Signup(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	UpdateNonSensitiveInfo(ctx context.Context,
		user domain.User) error
	FindById(ctx context.Context,
		uid int64) (domain.User, error)
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
}

type userService struct{
	repo repository.UserRepository
}

//初始化UserService
func NewUserService(repo repository.UserRepository) UserService{
	return &userService{
		repo: repo,
	}
}

func (svc *userService) Signup(ctx context.Context,u domain.User)error{
	//加密密码
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password),bcrypt.DefaultCost)
	if err != nil{
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx,u)
}

func (svc *userService) Login(ctx context.Context,email,password string)(domain.User,error){
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

func (svc *userService) FindById(ctx context.Context, uid int64) (domain.User,error) {
	return svc.repo.FindById(ctx,uid)
}

func (svc *userService) FindOrCreate(ctx context.Context, phone string) (domain.User,error){
	//先查找，假定大部分用户是已经存在的用户
	u, err := svc.repo.FindByPhone(ctx,phone)
	if err != repository.ErrUserNotFound {
		//两种情况
		//err == nil，u是可用的
		//err != nil, 系统错误
		return u, err
	}
	//用户没找到
	err = svc.repo.Create(ctx, domain.User{
		Phone: phone,
	})
	//两种可能，一种是err恰好是唯一索引冲突(phone)
	//一种是err != nil，系统错误
	if err!= nil && err != repository.ErrDuplicateUser{
		return domain.User{}, err
	}
	//要么err == nil，要么ErrDuplicatedUser，也代表用户存在
	//主从延迟问题，理论上讲，强制走主库
	return svc.repo.FindByPhone(ctx,phone)
}

func (svc *userService) UpdateNonSensitiveInfo(ctx context.Context, user domain.User) error {
	//更新nickname等字段
	return svc.repo.UpdateNonZeroFields(ctx, user)
}