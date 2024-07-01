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
	userID, _ := strconv.Atoi(c.PostForm("userID"))
	checklist := checklist{
		ID:     len(checklists) + 1,
		UserID: userID,
		Name:   c.PostForm("name"),
		Items:  []item{},
	}

	if checklist.Name == "" {
		c.JSON(400, gin.H{
			"message": "Name is required",
		})
	} else {
		checklists[checklist.ID] = checklist
		c.JSON(200, gin.H{
			"message":   "Checklist created",
			"checklist": checklist,
		})
	}
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
	checklistID, _ := strconv.Atoi(c.Param("id"))
	itemID, _ := strconv.Atoi(c.Param("itemID"))
	updatedItem := item{
		Text:    c.PostForm("text"),
		Checked: c.PostForm("checked") == "true",
	}

	list, ok := checklists[checklistID]

	updated := false
	if ok {
		for i, item := range list.Items {
			if item.ID == itemID {
				list.Items[i].Checked = updatedItem.Checked
				list.Items[i].Text = updatedItem.Text
				updated = true
			}
		}
	}

	if updated {
		checklists[checklistID] = list
		c.JSON(200, gin.H{
			"message":     "Item updated",
			"updatedItem": updatedItem,
		})
	} else {
		c.JSON(400, gin.H{
			"message": "No item found to update",
		})
	}
}
