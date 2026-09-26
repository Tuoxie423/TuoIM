package api

import (
	"github.com/gin-gonic/gin"

	"ginchat/utils"
)

// Index godoc
// @Summary      首页
// @Description  欢迎页
// @Tags         首页
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /index [get]
func Index(c *gin.Context) {
	utils.Success(c, gin.H{"message": "welcome!"})
}
