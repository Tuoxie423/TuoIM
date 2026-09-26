package utils

import (
	"errors"
	"ginchat/config"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWT struct {
	AccessTokenSecret  []byte
	RefreshTokenSecret []byte
}

var (
	TokenExpired     error = errors.New("token expired")
	TokenNotValidYet error = errors.New("not a active token yet")
	TokenMalformed   error = errors.New("that's not even a token") // JWT格式不正确
	TokenInvalid     error = errors.New("couldn't handle this token")
)

func NewJWT() *JWT {
	return &JWT{
		AccessTokenSecret:  []byte(config.Global.Jwt.AccessTokenSecret),
		RefreshTokenSecret: []byte(config.Global.Jwt.RefreshTokenSecret),
	}
}

func GetAccessToken(c *gin.Context) string {
	token := c.GetHeader("x-access-token")

	return token
}

func GetRefreshToken(c *gin.Context) string {
	refreshToken := c.GetHeader("x-refresh-token")
	return refreshToken
}

// ClearRefreshToken 清除Refresh Token的cookie
func ClearRefreshToken(c *gin.Context) {
	// 获取请求的host，如果失败则取原始请求host
	host, _, err := net.SplitHostPort(c.Request.Host)
	if err != nil {
		host = c.Request.Host
	}
	// 调用setCookie设置cookie值为空并过期，删除refresh-token
	setCookie(c, "x-refresh-token", "", -1, host)
}

func setCookie(c *gin.Context, name, value string, maxAge int, host string) {
	// 判断host是否是IP地址
	if net.ParseIP(host) != nil {
		// 如果是IP地址，设置cookie的domain为“/”
		c.SetCookie(name, value, maxAge, "/", "", false, true)
	} else {
		// 如果是域名，设置cookie的domain为域名
		c.SetCookie(name, value, maxAge, "/", host, false, true)
	}
}

func (j *JWT) CreateAccessToken(claims *Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.AccessTokenSecret)
}

func (j *JWT) CreateRefreshToken(claims *Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.RefreshTokenSecret)
}

// ParseAccessToken 解析并校验 access token
func (j *JWT) ParseAccessToken(tokenString string) (*Claims, error) {
	return j.parseToken(tokenString, j.AccessTokenSecret)
}

// ParseRefreshToken 解析并校验 refresh token
func (j *JWT) ParseRefreshToken(tokenString string) (*Claims, error) {
	return j.parseToken(tokenString, j.RefreshTokenSecret)
}

// parseToken 通用解析逻辑：验签 + 校验 claims，并把 jwt 错误映射成项目自己的错误
func (j *JWT) parseToken(tokenString string, secret []byte) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, TokenExpired
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, TokenNotValidYet
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, TokenMalformed
		default:
			return nil, TokenInvalid
		}
	}
	if !token.Valid {
		return nil, TokenInvalid
	}
	return claims, nil
}
