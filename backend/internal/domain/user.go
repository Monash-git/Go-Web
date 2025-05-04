// 放置领域对象
package domain

import "time"

//领域对象，非数据库对象，为业务对象
type User struct{
	Id       int64
	Email string
	Password string
	Phone string

	Nickname string
	// YYYY-MM-DD
	Birthday time.Time
	AboutMe  string

	Ctime time.Time
}