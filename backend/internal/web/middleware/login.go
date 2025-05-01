package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct{

}

func (m *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/signup" || path == "/users/login"{
			return
		}
		//session登录校验逻辑
		sess := sessions.Default(ctx)
		userId := sess.Get("userId")
		if  userId == nil {
			//中断，不执行后面的业务逻辑
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		//登录态刷新策略
		now := time.Now()
		//计算每个token已经度过的时间
		//获取上一次刷新时间
		const updateTimeKey = "update_time"
		val := sess.Get(updateTimeKey)
		lastUpdateTime, ok := val.(time.Time)
		if val == nil || !ok || now.Sub(lastUpdateTime) > time.Second*10 {
			sess.Set(updateTimeKey,now)
			sess.Set("userId",userId)
			err := sess.Save()
			if err != nil{
				//打日志
				fmt.Println(err)
			}
		}

	}
}