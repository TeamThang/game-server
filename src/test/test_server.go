package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "./src/test/ws_test.html")
	})

	r.Run(":5005")
}
