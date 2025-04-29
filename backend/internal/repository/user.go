package repository

import (
	"backend/internal/domain"
	"backend/internal/repository/dao"
	"context"
)

var ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail

type UserRepository struct{
	dao *dao.UserDAO
}

//初始化UserRepository
func NewUserRepository(d *dao.UserDAO) *UserRepository{
	return &UserRepository{
		dao: d,
	}
}

func (ur *UserRepository) Create(ctx context.Context, u domain.User) error{
	return ur.dao.Insert(ctx, dao.User{
		Email: u.Email,
		Password: u.Password,
	})
}