package utils

import (
	"errors"

	"ginchat/config"

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
