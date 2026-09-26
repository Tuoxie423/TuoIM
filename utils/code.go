package utils

// 业务错误码
const (
	// CodeSuccess 成功
	CodeSuccess = 0

	// 鉴权类错误码（401xx）
	CodeUnauthorized   = 40100 // 未登录
	CodeTokenExpired   = 40101 // 登录已过期
	CodeTokenInvalid   = 40102 // 凭证无效
	CodeTokenMalformed = 40103 // 凭证格式错误
)
