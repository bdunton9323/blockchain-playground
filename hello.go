package main

import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()

	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello World!")
	})

	router.Run(":3000")
}
