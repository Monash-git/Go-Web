package middleware

import (
	"backend/internal/web"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type LoginJWTMiddlewareBuilder struct{

}

func (m *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/signup" || path == "/users/login"{
			return
		}
		//jwt登录校验逻辑
		//根据约定：token在Authorization头部
		//Bearer xxxx
		authcode := ctx.GetHeader("Authorization")
		if authcode == ""{
			//没登陆，没有token，Authorization这个头部都没有
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		segs := strings.Split(authcode," ")
		if len(segs) != 2 {
			//没登陆，或者Authorization中的内容是乱传的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		tokenStr := segs[1]
		var uc web.UserClaims
		//将Bearer xxx中的xxx当作JWT，反序列化到uc中，并使用你自己的JWTkey验证
		token, err := jwt.ParseWithClaims(tokenStr,&uc,func(t *jwt.Token) (interface{}, error) {
			return web.JWTKey,nil
		})
		if err != nil{
			//token不对，token是伪造的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if token == nil || !token.Valid {
			//token解析出来了，但是token可能是非法的或者过期的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if uc.UserAgent != ctx.GetHeader("User-Agent"){
			//能进这个分支的，大概是攻击者
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		
		expireTime := uc.ExpiresAt
		//剩余过期时间<50s刷新
		if time.Until(expireTime.Time) < 50*time.Second {
			uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute*5))
			tokenStr,err := token.SignedString(web.JWTKey)
			ctx.Header("x-jwt-token",tokenStr)
			if err != nil {
				//这里不能中断，因为仅仅是过期时间没有刷新，但是用户是已经登陆的
				log.Println(err)
			}
		}
		ctx.Set("user",uc)
	}
}