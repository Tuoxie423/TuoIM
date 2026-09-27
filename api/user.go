package api

import (
	"github.com/gin-gonic/gin"

	"ginchat/service"
	"ginchat/utils"
)

var userService = new(service.UserService)

// RegisterRequest 注册请求参数
type RegisterRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// GetUser godoc
// @Summary      获取用户列表
// @Tags         用户
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /getuser [get]
func GetUser(c *gin.Context) {
	data, err := userService.GetUser()
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, data)
}

// Register godoc
// @Summary      用户注册
// @Description  通过手机号注册新用户
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        body body RegisterRequest true "注册参数"
// @Success      200 {object} utils.Response
// @Router       /user/register [post]
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}
	user, err := userService.Register(req.Name, req.Password, req.Phone)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}
	utils.Success(c, user)
}

// Login godoc
// @Summary      用户登录
// @Description  手机号 + 密码登录
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        body body LoginRequest true "登录参数"
// @Success      200 {object} utils.Response
// @Router       /user/login [post]
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, 400, "参数错误: "+err.Error())
		return
	}
	user, err := userService.Login(req.Phone, req.Password)
	if err != nil {
		utils.Error(c, utils.CodeUnauthorized, err.Error())
		return
	}
	j := utils.NewJWT()
	claims := utils.CreateAccessClaims(int64(user.ID)) // user.ID 是 uint，要转 int64
	token, err := j.CreateAccessToken(claims)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, gin.H{"token": token, "user": user})
}
