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
	bizLogin = "login"
)

type UserHandler struct{
	//邮箱和密码的校验
	emailRegexExp *regexp.Regexp
	passwordRegexExp *regexp.Regexp
	svc service.UserService
	codeSvc service.CodeService
}

type UserClaims struct{
	jwt.RegisteredClaims
	uid int64
	UserAgent string
}

//UserHandler初始化方法
func NewUserHandler(svc service.UserService, codeSvc service.CodeService) *UserHandler{
	return &UserHandler{
		svc: svc,
		codeSvc: codeSvc,
		emailRegexExp: regexp.MustCompile(emailRegexPattern,regexp.None),
		passwordRegexExp: regexp.MustCompile(passwordRegexPattern,regexp.None),
	}
}

//路由注册
func (h *UserHandler) RegisterRoutes(server *gin.Engine){
	//分组路由
	ug := server.Group("/users")
	ug.POST("/signup",h.Signup)
	ug.POST("/login",h.LoginJWT)
	ug.POST("/edit",h.Edit)
	ug.POST("/profile",h.Profile)

	//手机验证码登录相关功能
	ug.POST("/login_sms/code/send",h.SendSMSLoginCode)
	ug.POST("/login_sms", h.LoginSMS)
}

func (c *UserHandler) SendSMSLoginCode(ctx *gin.Context){
	//定义一个结构体接收数据
	type SignUpReq struct{
		Phone string `json:"phone"`
	}

	var req SignUpReq

	//使用Bind方法接受请求，如果有误会直接返回响应到前端
	if err := ctx.Bind(&req); err != nil{
		return
	}
	//这里可以校验Req
	if req.Phone == ""{
		ctx.JSON(http.StatusOK,Result{
			Code: 4,
			Msg: "请输入手机号",
		})
		return
	}
	err := c.codeSvc.Send(ctx,bizLogin,req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, Result{
			Msg: "发送成功",
		})
	case service.ErrCodeSendTooMany:
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg:  "短信发送太频繁，请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		//补日志
	}

}

//短信验证登录
func (h *UserHandler) LoginSMS(ctx *gin.Context){
	//定义一个结构体接收数据
	type SignUpReq struct{
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	var req SignUpReq

	//使用Bind方法接受请求，如果有误会直接返回响应到前端
	if err := ctx.Bind(&req); err != nil{
		return
	}

	ok,err := h.codeSvc.Verify(ctx, bizLogin,req.Phone, req.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg: "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Result{
			Code: 4,
			Msg: "验证码错误，请重新输入",
		})
		return
	}
	u, err := h.svc.FindOrCreate(ctx,req.Phone)
	if err != nil{
		ctx.JSON(http.StatusOK,Result{
			Code: 5,
			Msg: "系统错误",
		})
	}
	h.setJWTToken(ctx,u.Id)
	ctx.JSON(http.StatusOK,Result{
		Msg: "登陆成功",
	})
}



//SignUp用户注册接口
func (h *UserHandler) Signup(ctx *gin.Context){
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
	isEmail, err := h.emailRegexExp.MatchString(req.Email)
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

	isPassword, err := h.passwordRegexExp.MatchString(req.Password)
	if err != nil {
		ctx.String(http.StatusOK,"系统错误")
		return
	}
	if !isPassword {
		ctx.String(http.StatusOK,"密码必须包含字母、数字、特殊字符，并且不少于八位")
		return
	} 
	//校验通过，写入数据库
	err = h.svc.Signup(ctx.Request.Context(),domain.User{
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

func(h *UserHandler) setJWTToken(ctx *gin.Context, uid int64){
	uc := UserClaims{
		uid: uid,
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
}

func (h *UserHandler) LoginJWT(ctx *gin.Context){
	type Req struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil{
		return
	}
	u, err := h.svc.Login(ctx, req.Email,req.Password)
	//JWT逻辑
	switch err{
	case nil:
		h.setJWTToken(ctx,u.Id)
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或者密码错误")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

//session登录
func (h *UserHandler) Login(ctx *gin.Context){
	type Req struct{
		Email string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil{
		return
	}
	u, err := h.svc.Login(ctx, req.Email,req.Password)
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


func (h *UserHandler) Edit(ctx *gin.Context){
	// 嵌入一段刷新过期时间的代码
	type Req struct {
		// 改邮箱，密码，或者能不能改手机号

		Nickname string `json:"nickname"`
		// YYYY-MM-DD
		Birthday string `json:"birthday"`
		AboutMe  string `json:"aboutMe"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	//sess := sessions.Default(ctx)
	//sess.Get("uid")
	uc, ok := ctx.MustGet("user").(UserClaims)
	if !ok {
		//ctx.String(http.StatusOK, "系统错误")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	// 用户输入不对
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		//ctx.String(http.StatusOK, "系统错误")
		ctx.String(http.StatusOK, "生日格式不对")
		return
	}
	err = h.svc.UpdateNonSensitiveInfo(ctx, domain.User{
		Id:       uc.uid,
		Nickname: req.Nickname,
		Birthday: birthday,
		AboutMe:  req.AboutMe,
	})
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	ctx.String(http.StatusOK, "更新成功")
}

func (h *UserHandler) Profile(ctx *gin.Context){
	//us := ctx.MustGet("user").(UserClaims)
	//ctx.String(http.StatusOK, "这是 profile")
	// 嵌入一段刷新过期时间的代码

	uc, ok := ctx.MustGet("user").(UserClaims)
	if !ok {
		//ctx.String(http.StatusOK, "系统错误")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	u, err := h.svc.FindById(ctx, uc.uid)
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	type User struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		AboutMe  string `json:"aboutMe"`
		Birthday string `json:"birthday"`
	}
	ctx.JSON(http.StatusOK, User{
		Nickname: u.Nickname,
		Email:    u.Email,
		AboutMe:  u.AboutMe,
		Birthday: u.Birthday.Format(time.DateOnly),
	})
}