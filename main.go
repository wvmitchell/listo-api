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
	r.POST("/checklist/:id/item", postItem)
	r.PUT("/checklist/:id/item/:itemID", putItem)

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

func postItem(c *gin.Context) {
	checklistID, _ := strconv.Atoi(c.Param("id"))
	item := item{
		Text:    c.PostForm("text"),
		Checked: false,
	}
	list, ok := checklists[checklistID]
	if ok {
		list.Items = append(list.Items, item)
		checklists[checklistID] = list
		c.JSON(200, gin.H{
			"message": "Item added",
		})
	} else {
		c.JSON(404, gin.H{
			"message": "Could not add item",
		})
	}
}

func putItem(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Item updated",
	})
}
