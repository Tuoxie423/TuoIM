package middleware

import (
	"errors"
	"ginchat/config/global"
	"ginchat/utils"

	"strconv"

	"github.com/gin-gonic/gin"
)

// JWTAuth 是一个中间件函数，验证请求中的JWT token是否合法
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求中的Access Token和Refresh Token
		accessToken := utils.GetAccessToken(c)
		refreshToken := utils.GetRefreshToken(c)

		// 检查Refresh Token是否在黑名单中，如果是，则清除Refresh Token并返回未授权错误
		if jwtService.IsInBlacklist(refreshToken) {
			utils.ClearRefreshToken(c)
			utils.Response("Account logged in from another location or token is invalid", c)
			c.Abort() // 终止请求的后续处理
			return
		}

		// 创建一个JWT实例，用于后续的token解析与验证
		j := utils.NewJWT()

		// 解析Access Token
		claims, err := j.ParseAccessToken(accessToken)
		if err != nil {
			// 如果解析失败并且Access Token为空或过期
			if accessToken == "" || errors.Is(err, utils.TokenExpired) {
				// 尝试解析Refresh Token
				refreshClaims, err := j.ParseRefreshToken(refreshToken)
				if err != nil {
					// 如果Refresh Token也无法解析，清除Refresh Token并返回未授权错误
					utils.ClearRefreshToken(c)
					utils.Response.NoAuth("Refresh token expired or invalid", c)
					c.Abort()
					return
				}

				// 如果Refresh Token有效，通过其UserID获取用户信息
				var user database.User
				if err := global.DB.Select("uuid", "role_id").Take(&user, refreshClaims.UserID).Error; err != nil {
					// 如果没有找到该用户，清除Refresh Token并返回未授权错误
					utils.ClearRefreshToken(c)
					response.NoAuth("The user does not exist", c)
					c.Abort()
					return
				}

				// 使用Refresh Token的用户信息创建一个新的Access Token的Claims
				newAccessClaims := utils.CreateAccessClaims(request.BaseClaims{
					UserID: refreshClaims.UserID,
					UUID:   user.UUID,
					RoleID: user.RoleID,
				})
				// 创建新的Access Token
				newAccessToken, err := j.CreateAccessToken(newAccessClaims)
				if err != nil {
					// 如果生成新的Access Token失败，清除Refresh Token并返回未授权错误
					utils.ClearRefreshToken(c)
					response.NoAuth("Failed to create new access token", c)
					c.Abort()
					return
				}

				// 将新的Access Token和过期时间添加到响应头中
				c.Header("new-access-token", newAccessToken)
				c.Header("new-access-expires-at", strconv.FormatInt(newAccessClaims.ExpiresAt.Unix(), 10))

				// 将新的Claims信息存入Context，供后续使用
				c.Set("claims", &newAccessClaims)
				c.Next() // 继续后续的处理
				return
			}

			// 如果Access Token无效且不满足刷新条件，清除Refresh Token并返回未授权错误
			utils.ClearRefreshToken(c)
			response.NoAuth("Invalid access token", c)
			c.Abort()
			return
		}

		// 如果Access Token合法，将其Claims信息存入Context
		c.Set("claims", claims)
		c.Next() // 继续后续的处理
	}
}
