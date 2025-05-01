package main

import (
	"backend/internal/config"
	"backend/internal/repository"
	"backend/internal/repository/dao"
	"backend/internal/service"
	"backend/internal/web"
	"backend/internal/web/middleware"
	"backend/pkg/ginx/middleware/ratelimit"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	redisStore "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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

		AllowHeaders: []string{"Content-Type","Authorization"},
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

	redisClient := redis.NewClient(&redis.Options{
		Addr: config.Config.Redis.Addr,
	})

	server.Use(ratelimit.NewBuilder(redisClient,time.Second,1).Build())

	// useSeesion(server)
	useJWT(server)

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

//使用jwt实现登录态保持
func useJWT(server *gin.Engine){
	login := middleware.LoginJWTMiddlewareBuilder{}
	server.Use(login.CheckLogin())
}

//使用session实现登录态保持
func useSeesion (server *gin.Engine){
	//登录校验
	login := &middleware.LoginMiddlewareBuilder{}
	//存储数据，也就是userId存在哪里
	//直接存cookie，名字为ssid
	// store := cookie.NewStore([]byte("secret"))
	//基于redis的实现
	store, err := redisStore.NewStore(
		16,
		"tcp",
		"localhost:6379",
		"",
		"",                                       // 如果 Redis 没有密码，就传空字符串
		[]byte("k6CswdUm75WKcbM68UQUuxVsHSpTCwgK"), // authentication key
		[]byte("k6CswdUm75WKcbM68UQUuxVsHSpTCwgA"), // encryption key
	)
	if err != nil {
		panic(err)
	}

	server.Use(sessions.Sessions("ssid",store),login.CheckLogin())
}