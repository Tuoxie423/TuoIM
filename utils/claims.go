package utils

import (
	"time"

	"ginchat/config"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 自定义 JWT 声明，内嵌标准声明（含过期时间等）
type Claims struct {
	jwt.RegisteredClaims
	Type   string `json:"type"`    // token 类型：access / refresh
	UserID int64  `json:"user_id"` // 用户 ID
}

// CreateAccessClaims 构造 access token 的声明
func CreateAccessClaims(userID int64) *Claims {
	now := time.Now()
	return &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(config.Global.Jwt.AccessExpireTime) * time.Second)),
		},
		Type:   "access",
		UserID: userID,
	}
}

// CreateRefreshClaims 构造 refresh token 的声明
func CreateRefreshClaims(userID int64) *Claims {
	now := time.Now()
	return &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(config.Global.Jwt.RefreshExpireTime) * time.Second)),
		},
		Type:   "refresh",
		UserID: userID,
	}
}
