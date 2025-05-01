package ratelimit

//滑动窗口限流中间件
import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Builder struct{
	prefix string
	cmd redis.Cmdable //Redis 客户端接口
	interval time.Duration
	//阈值
	rate int
}

//go:embed slide_window.lua
var luaScript string

func NewBuilder(cmd redis.Cmdable, interval time.Duration, rate int) *Builder{
	return &Builder{
		cmd: cmd,
		prefix: "ip-limiter",
		interval: interval,
		rate: rate,
	}
}

func (b *Builder) Prefix(prefix string) *Builder{
	b.prefix = prefix
	return b
}

func (b *Builder) limit(ctx *gin.Context) (bool,error){
	key := fmt.Sprintf("%s:%s",b.prefix,ctx.ClientIP())
	return b.cmd.Eval(ctx,luaScript,[]string{key},b.interval.Milliseconds(),b.rate,time.Now().UnixMilli()).Bool()
}

func (b *Builder) Build() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		limited, err := b.limit(ctx)
		if err != nil {
			log.Println(err)
			//如果Redis出错，这里有两种策略：
			//保守：直接返回500，全部拒绝
			ctx.AbortWithStatus(http.StatusInternalServerError)
			//激进：放行所有请求，可能会导致数据库无法承载过多的请求崩溃
			//ctx.Next()
			return
		}
		if limited {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
} 