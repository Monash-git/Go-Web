package main

import (
	"backend/internal/repository"
	"backend/internal/repository/dao"
	"backend/internal/service"
	"backend/internal/web"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main(){
	db := initDB()
	
	server := initWebServer()
	initUser(server,db)
	server.Run(":8080")
}

func initWebServer() *gin.Engine{
	server:= gin.Default()

	//跨域问题
	server.Use(cors.New(cors.Config{
		//AllowAllOrigins: true,
		//AllowOrigins:     []string{"http://localhost:3000"},
		AllowCredentials: true,

		AllowHeaders: []string{"Content-Type"},
		//AllowHeaders: []string{"content-type"},
		//AllowMethods: []string{"POST"},
		AllowOriginFunc: func(origin string) bool {
			if strings.HasPrefix(origin, "http://localhost") {
				//if strings.Contains(origin, "localhost") {
				return true
			}
			return strings.Contains(origin, "your_company.com")
		},
		MaxAge: 12 * time.Hour,
	}))
	return server
}

func initUser(server *gin.Engine, db *gorm.DB){
	ud := dao.NewUserDAO(db)
	ur := repository.NewUserRepository(ud)
	us := service.NewUserService(ur)
	c := web.NewUserHandler(us)
	c.RegisterRoutes(server)
}

func initDB() *gorm.DB{
	db, err := gorm.Open(mysql.Open("root:root@tcp(localhost:13316)/goweb"))
	if err != nil{
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil{
		panic(err)
	}
	return db
}