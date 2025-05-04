package repository

import (
	"backend/internal/domain"
	"backend/internal/repository/cache"
	"backend/internal/repository/dao"
	"context"
	"database/sql"
	"log"
)

var (
	ErrDuplicateUser= dao.ErrDuplicateEmail
	ErrUserNotFound = dao.ErrRecordNotFound
)

type UserRepository interface{
	Create(ctx context.Context, u domain.User) error
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	UpdateNonZeroFields(ctx context.Context, user domain.User) error
	FindByPhone(ctx context.Context, phone string) (domain.User, error)
	FindById(ctx context.Context, uid int64) (domain.User, error)
}

type CachedUserRepository struct{
	dao dao.UserDAO
	cache cache.UserCache
}

//初始化UserRepository
func NewCachedUserRepository(d dao.UserDAO, c cache.UserCache) UserRepository{
	return &CachedUserRepository{
		dao: d,
		cache: c,
	}
}

func (repo *CachedUserRepository) UpdateNonZeroFields(ctx context.Context, user domain.User) error {
	return repo.dao.UpdateById(ctx,repo.toEntity(user))
}

func (repo *CachedUserRepository) Create(ctx context.Context, u domain.User) error{
	return repo.dao.Insert(ctx, repo.toEntity(u))
}

//根据email进行用户查询
func (repo *CachedUserRepository) FindByEmail(ctx context.Context, email string)(domain.User,error){
	u, err:= repo.dao.FindByEmail(ctx,email)
	if err != nil{
		return domain.User{},err
	}
	return repo.toDomain(u),nil
}

//将dao的User对象转换为domain的领域对象
func (repo *CachedUserRepository) toDomain(u dao.User) domain.User{
	return domain.User{
		Id: u.Id,
		Email: u.Email.String,
		Password: u.Password,
	}
}

func (repo *CachedUserRepository) toEntity(u domain.User) dao.User{
	return dao.User{
		Id: u.Id,
		Email: sql.NullString{
			String: u.Email,
			Valid:  u.Email != "",
		},
		Phone: sql.NullString{
			String: u.Phone,
			Valid:  u.Phone != "",
		},
		Password: u.Password,
	}
}

func (repo *CachedUserRepository) FindById(ctx context.Context, uid int64) (domain.User, error){
	du, err := repo.cache.Get(ctx, uid)
	if err != nil {
		return	du, nil
	}
	
	//err不为nil，就要查询数据库
	//err有两种可能
	//1、key不存在，说明redis是正常的
	//2、访问redis有问题，可能是网络有问题，也可能是redis本身崩溃的

	u, err := repo.dao.FindById(ctx,uid)
	if err != nil {
		return domain.User{}, err
	}
	du = repo.toDomain(u)
	err = repo.cache.Set(ctx,du)
	if err != nil {
		//网络崩了，或是redis崩了
		log.Println(err)
	}
	return du, nil
}

func (repo *CachedUserRepository) FindByPhone(ctx context.Context, phone string) (domain.User, error){
	u, err := repo.dao.FindByPhone(ctx,phone)
	if err != nil {
		return domain.User{}, err
	}
	return repo.toDomain(u), nil
}