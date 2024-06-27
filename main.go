// Package main is the entry point of the application.
package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	items := []string{}

	r.GET("/items", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"items": items,
		})
	})

	r.POST("/items", func(c *gin.Context) {
		item := c.PostForm("item")
		items = append(items, item)
		c.JSON(200, gin.H{
			"message": "Item added successfully",
		})
	})

	e := r.Run(":7235")

	if e != nil {
		panic(e)
	}
}
