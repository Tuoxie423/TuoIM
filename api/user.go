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

// Register godoc
// @Summary      用户注册
// @Description  通过手机号注册新用户
// @Tags         用户
// @Accept       json
// @Produce      json
// @Param        body body RegisterRequest true "注册参数"
// @Success      200 {object} utils.Response
// @Router       /register [post]
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
// @Router       /login [post]
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

// GetUserInfo godoc
// @Summary      获取当前登录用户信息
// @Description  通过 token 解析出 userID 并返回用户信息
// @Tags         用户
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Success      200 {object} utils.Response
// @Router       /users/me [get]
func GetUserInfo(c *gin.Context) {
	// 从 context 取出中间件存的 claims
	claimsAny, ok := c.Get("claims")
	if !ok {
		utils.Error(c, utils.CodeUnauthorized, "未登录")
		return
	}
	userClaims, ok := claimsAny.(*utils.Claims)
	if !ok {
		utils.Error(c, 500, "凭证解析失败")
		return
	}

	// 用 claims 里的 userID 查完整用户信息
	user, err := userService.GetUserByID(userClaims.UserID)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, user)
}
