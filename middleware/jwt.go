package middleware

import (
	"errors"

	"ginchat/utils"

	"github.com/gin-gonic/gin"
)

// JWTAuth 校验请求中的 access token，失败返回未授权错误
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := utils.NewJWT().ParseAccessToken(utils.GetAccessToken(c))
		if err != nil {
			switch {
			case errors.Is(err, utils.TokenExpired):
				utils.Error(c, utils.CodeTokenExpired, "登录已过期，请重新登录")
			default:
				utils.Error(c, utils.CodeTokenInvalid, "无效的凭证")
			}
			c.Abort()
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}
