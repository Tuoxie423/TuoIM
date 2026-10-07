package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ginchat/utils"

	"github.com/gin-gonic/gin"
)

// allowedImageExt 允许上传的图片扩展名
var allowedImageExt = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
}

// allowedImageMime 允许上传的图片 MIME 类型（与 http.DetectContentType 的结果匹配）
var allowedImageMime = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// Upload godoc
// @Summary      上传图片
// @Tags         文件
// @Accept       multipart/form-data
// @Produce      json
// @Param        Authorization header string true "Bearer token"
// @Param        file formData file true "图片文件"
// @Success      200 {object} utils.Response
// @Router       /file/upload [post]
func Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, 400, "文件上传失败")
		return
	}

	// 扩展名校验：快速拒绝明显非图片的文件
	if ext := strings.ToLower(filepath.Ext(file.Filename)); !allowedImageExt[ext] {
		utils.Error(c, 400, "仅支持 jpg/jpeg/png/gif/webp 格式的图片")
		return
	}

	// 内容嗅探校验：防止把非图片文件改名成 .jpg 上传
	f, err := file.Open()
	if err != nil {
		utils.Error(c, 400, "文件读取失败")
		return
	}
	head := make([]byte, 512)
	n, _ := f.Read(head)
	f.Close()
	if !allowedImageMime[http.DetectContentType(head[:n])] {
		utils.Error(c, 400, "文件内容不是图片")
		return
	}

	// 确保 upload 目录存在
	if err := os.MkdirAll("./upload", 0755); err != nil {
		utils.Error(c, 500, "创建目录失败")
		return
	}

	// 唯一文件名，防止覆盖
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := "./upload/" + filename

	if err := c.SaveUploadedFile(file, dst); err != nil {
		utils.Error(c, 500, "保存失败")
		return
	}

	utils.Success(c, gin.H{"url": "/upload/" + filename})
}
