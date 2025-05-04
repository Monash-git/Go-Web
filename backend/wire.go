//go:build wireinject

package main

import (
	"backend/internal/repository"
	"backend/internal/repository/cache"
	"backend/internal/repository/dao"
	"backend/internal/service"
	"backend/internal/web"
	"backend/ioc"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		//第三方依赖
		ioc.InitRedis, ioc.InitDB,
		//DAO部分
		dao.NewUserDAO,

		//cache部分
		cache.NewCodeCache, cache.NewUserCache,
		// repository 部分
		repository.NewCachedUserRepository,
		repository.NewCodeRepository,

		// Service 部分
		ioc.InitSMSService,
		service.NewUserService,
		service.NewCodeService,

		// handler 部分
		web.NewUserHandler,

		ioc.InitGinMiddlewares,
		ioc.InitWebServer,
	)
	return gin.Default()
}