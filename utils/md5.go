package utils

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5 返回字符串的 md5 摘要（16 进制小写）
func MD5(s string) string {
	sum := md5.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}
