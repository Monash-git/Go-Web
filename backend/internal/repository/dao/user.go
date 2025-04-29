package dao

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var ErrUserDuplicateEmail = errors.New("邮箱冲突")

//数据库对象，直接映射到数据库的表中
type User struct{
	//自增，主键
	Id int64 `gorm:"primaryKey,autoIncrement"`
	//唯一索引
	Email string `gorm:"unique"`
	Password string

	//创建时间
	Ctime int64
	//更新时间
	Utime int64
}

type UserDAO struct{
	db *gorm.DB
}

//初始化UserDAO，依赖注入
func NewUserDAO(db *gorm.DB) *UserDAO{
	return &UserDAO{
		db: db,
	}
}

func (ud *UserDAO) Insert(ctx context.Context, u User)error{
	//写入数据库
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := ud.db.WithContext(ctx).Create(&u).Error
	//获取数据库的唯一索引冲突错误
	if me,ok := err.(*mysql.MySQLError); ok {
		const uniqueIndexErrNo uint16 = 1062
		if me.Number == uniqueIndexErrNo {
			return ErrUserDuplicateEmail
		}
	}
	return err
}