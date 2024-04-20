package framework

import (
	"github.com/gin-gonic/gin"
)

func Setup() {
	r := gin.Default()

	r.POST("/post", func(c *gin.Context) {
		text := c.PostForm("text")
		c.JSON(200, gin.H{"text": text})
	})

	r.Run() //启动服务 默认端口8080
}
