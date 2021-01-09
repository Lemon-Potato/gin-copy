package main

import "gin_copy/gin"

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Content) {
		c.String(200, "pong")
	})
	r.Run(":8082")
}