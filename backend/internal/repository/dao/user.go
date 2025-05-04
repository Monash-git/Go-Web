package dao

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var (
	ErrDuplicateEmail = errors.New("邮箱冲突")
	//未找到用户
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO interface{
	Insert(ctx context.Context, u User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	UpdateById(ctx context.Context, entity User) error
	FindById(ctx context.Context, uid int64) (User, error)
	FindByPhone(ctx context.Context, phone string) (User, error)
}

//数据库对象，直接映射到数据库的表中
type User struct{
	//自增，主键
	Id int64 `gorm:"primaryKey,autoIncrement"`
	//唯一索引,同时可以为Null
	Email sql.NullString `gorm:"unique"`
	Password string
	//可以为Null的列
	Phone sql.NullString `gorm:"unique"`

	Nickname string `gorm:"type=varchar(128)"`
	// YYYY-MM-DD
	Birthday int64
	AboutMe  string `gorm:"type=varchar(4096)"`

	//创建时间
	Ctime int64
	//更新时间
	Utime int64
}

type GORMUserDAO  struct{
	db *gorm.DB
}

//初始化UserDAO，依赖注入
func NewUserDAO(db *gorm.DB) UserDAO{
	return &GORMUserDAO{
		db: db,
	}
}

func (dao *GORMUserDAO) Insert(ctx context.Context, u User)error{
	//写入数据库
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	//获取数据库的唯一索引冲突错误
	if me,ok := err.(*mysql.MySQLError); ok {
		const uniqueIndexErrNo uint16 = 1062
		if me.Number == uniqueIndexErrNo {
			//邮箱冲突，用户冲突
			return ErrDuplicateEmail
		}
	}
	return err
}

func (dao *GORMUserDAO) UpdateById(ctx context.Context, entity User) error {
	//依赖于GORM的零值和主键更新特性
	//Update 非零值 WHERE id = ？
	return dao.db.WithContext(ctx).Model(&entity).Where("id=?",entity.Id).
		Updates(map[string]any{
			"utime":    time.Now().UnixMilli(),
			"nickname": entity.Nickname,
			"birthday": entity.Birthday,
			"about_me": entity.AboutMe,
		}).Error
}

func (dao *GORMUserDAO) FindByEmail(ctx context.Context, email string)(User, error){
	var u User
	err := dao.db.WithContext(ctx).Where("email=?",email).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) FindById(ctx context.Context, uid int64)(User,error){
	var u User
	err := dao.db.WithContext(ctx).Where("id=?",uid).First(&u).Error
	return u, err
}

func (dao *GORMUserDAO) FindByPhone(ctx context.Context, phone string) (User,error) {
	var res User
	err := dao.db.WithContext(ctx).Where("phone=?",phone).First(&res).Error
	return res, err
}

