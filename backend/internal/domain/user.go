package domain

import "time"

//领域对象，非数据库对象，为业务对象
type User struct{
	Email string
	Password string
	Ctime time.Time
}