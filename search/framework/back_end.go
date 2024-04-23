package framework

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Request struct {
	Text string `json:"text" binding:"required"`
}

type Answer struct {
	Score int    `json:"score"`
	Url   string `json:"url"`
}

type Responce struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg"`
	Data []Answer `json:"data"`
}

func Setup() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Main website",
		})
	})

	r.POST("/search", func(c *gin.Context) {
		var request Request
		var responce Responce
		err := c.ShouldBindJSON(&request)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		responce.Code = http.StatusOK
		responce.Msg = "查询成功"
		responce.Data = append(responce.Data, Answer{22, request.Text}, Answer{11, request.Text})
		c.JSON(http.StatusOK, responce)
		return
	})

	r.Run(":8080") //启动服务 默认端口8080
}
