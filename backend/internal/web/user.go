package web

import (
	"backend/internal/domain"
	"backend/internal/service"
	"net/http"
	"time"

	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/golang-jwt/jwt/v5"

	"github.com/gin-gonic/gin"
)

var ErrUserDuplicateEmail = service.ErrUserDuplicateEmail

var JWTKey = []byte("k6CswdUm77WKcbM68UQUuxVsHSpTCwgK")

//校验正则表达式
const(
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
)

type UserHandler struct{
	//邮箱和密码的校验
	emailRegexExp *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	svc *service.UserService
}

type UserClaims struct{
	jwt.RegisteredClaims
	uid int64
	UserAgent string
}

//UserHandler初始化方法
func NewUserHandler(svc *service.UserService) *UserHandler{
	return &UserHandler{
		svc: svc,
		emailRegexExp: regexp.MustCompile(emailRegexPattern,regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern,regexp.None),
	}
}

//路由注册
func (c *UserHandler) RegisterRoutes(server *gin.Engine){
	//分组路由
	ug := server.Group("/users")
	ug.POST("/signup",c.Signup)
	ug.POST("/login",c.LoginJWT)
	ug.POST("/edit",c.Edit)
	ug.POST("/profile",c.Profile)
}

//SignUp用户注册接口
func (c *UserHandler) Signup(ctx *gin.Context){
	//定义一个结构体接收数据
	type SignUpReq struct{
		Email string `json:"email"`
		Password string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	var req SignUpReq

	//使用Bind方法接受请求，如果有误会直接返回响应到前端
	if err := ctx.Bind(&req); err != nil{
		return
	}

	//校验逻辑
	isEmail, err := c.emailRegexExp.MatchString(req.Email)
	if err != nil {
		ctx.String(http.StatusOK,"系统错误")
		return
	}
	if !isEmail {
		ctx.String(http.StatusOK,"邮箱格式不正确")
		return
	}
	
	if req.Password != req.ConfirmPassword{
		ctx.String(http.StatusOK,"两次输入的密码不一致")
		return
	}

	isPassword, err := c.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK,"系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK,"密码必须包含字母、数字、特殊字符，并且不少于八位")
		return
	} 
	//校验通过，写入数据库
	err = c.svc.Signup(ctx.Request.Context(),domain.User{
		Email: req.Email,
		Password: req.Password,
	})
	switch err{
	case nil:
		ctx.String(http.StatusOK,"注册成功")
	case service.ErrUserDuplicateEmail:
		ctx.String(http.StatusOK,"邮箱已注册，请更换一个邮箱")
	default:
		ctx.String(http.StatusOK,"系统错误")
	}
}

//JWT登录
func (c *UserHandler) LoginJWT(ctx *gin.Context){
	type Req struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil{
		return
	}
	u, err := c.svc.Login(ctx, req.Email,req.Password)
	//JWT逻辑
	switch err{
	case nil:
		uc := UserClaims{
			uid: u.Id,
			UserAgent: ctx.GetHeader("User-Agent"),
			RegisteredClaims: jwt.RegisteredClaims{
				//五分钟过期
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute*5)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS512,uc)
		tokenStr,err := token.SignedString(JWTKey)
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
		}
		ctx.Header("x-jwt-token",tokenStr)
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或者密码错误")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

//session登录
func (c *UserHandler) Login(ctx *gin.Context){
	type Req struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil{
		return
	}
	u, err := c.svc.Login(ctx, req.Email,req.Password)
	//session逻辑
	switch err {
	case nil:
		sess := sessions.Default(ctx)
		sess.Set("userId",u.Id)
		sess.Options(sessions.Options{
			//60秒过期
			MaxAge: 60,
		})
		err = sess.Save()
		if err != nil {
			ctx.String(http.StatusOK,"系统错误")
		}
		ctx.String(http.StatusOK,"登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK,"用户名或密码错误")
	default:
		ctx.String(http.StatusOK,"系统错误")
	}
}


func (c *UserHandler) Edit(ctx *gin.Context){}
func (c *UserHandler) Profile(ctx *gin.Context){
	ctx.String(http.StatusOK, "这是 profile")
}