package repository

import (
	"backend/internal/domain"
	"backend/internal/repository/dao"
	"context"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound = dao.ErrUserNotFound
)

type UserRepository struct{
	dao *dao.UserDAO
}

//初始化UserRepository
func NewUserRepository(d *dao.UserDAO) *UserRepository{
	return &UserRepository{
		dao: d,
	}
}

func (repo *UserRepository) Create(ctx context.Context, u domain.User) error{
	return repo.dao.Insert(ctx, dao.User{
		Email: u.Email,
		Password: u.Password,
	})
}

//根据email进行用户查询
func (repo *UserRepository) FindByEmail(ctx context.Context, email string)(domain.User,error){
	u, err:= repo.dao.FindByEmail(ctx,email)
	if err != nil{
		return domain.User{},err
	}
	return repo.toDomain(u),nil
}

//将dao的User对象转换为domain的领域对象
func (repo *UserRepository) toDomain(u dao.User) domain.User{
	return domain.User{
		Id: u.Id,
		Email: u.Email,
		Password: u.Password,
	}
}