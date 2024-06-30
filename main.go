// Package main is the entry point of the application.
package main

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

func main() {
	r := gin.Default()

	r.GET("/checklist/:id", getChecklist)
	r.POST("/checklist", postChecklist)

	e := r.Run()

	if e != nil {
		panic(e)
	}
}

func getChecklist(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	list, ok := checklists[id]

	if ok {
		c.JSON(200, gin.H{
			"checklist": list,
		})
	} else {
		c.JSON(404, gin.H{
			"message": "Checklist not found",
		})
	}
}

func postChecklist(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Checklist created",
	})
}
