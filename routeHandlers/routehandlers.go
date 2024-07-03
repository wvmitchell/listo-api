// Package routehandlers provides the route handlers for the application.
package routehandlers

import (
	"checklist-api/models"
	"github.com/gin-gonic/gin"
	"strconv"
)

// GetChecklists returns all checklists for a user.
func GetChecklists(c *gin.Context) {
	userID, e := strconv.Atoi(c.GetHeader("userID"))
	userChecklists := []models.Checklist{}

	for _, list := range models.Checklists {
		if list.UserID == userID {
			userChecklists = append(userChecklists, list)
		}
	}

	if e != nil {
		c.JSON(400, gin.H{
			"message": "Invalid user ID",
		})
	} else {
		c.JSON(200, gin.H{
			"checklists": userChecklists,
		})
	}
}

// GetChecklist returns a single checklist.
func GetChecklist(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	list, ok := models.Checklists[id]

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

// PostChecklist creates a new checklist.
func PostChecklist(c *gin.Context) {
	userID, _ := strconv.Atoi(c.GetHeader("userID"))
	checklist := models.Checklist{
		ID:     len(models.Checklists) + 1,
		UserID: userID,
		Name:   c.PostForm("name"),
		Items:  []models.ChecklistItem{},
	}

	if checklist.Name == "" {
		c.JSON(400, gin.H{
			"message": "Name is required",
		})
	} else {
		models.Checklists[checklist.ID] = checklist
		c.JSON(200, gin.H{
			"message":   "Checklist created",
			"checklist": checklist,
		})
	}
}

// PostItem adds an item to a checklist.
func PostItem(c *gin.Context) {
	checklistID, _ := strconv.Atoi(c.Param("id"))
	var newItem models.ChecklistItem

	if err := c.BindJSON(&newItem); err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid request",
		})
		return
	}

	list, ok := models.Checklists[checklistID]
	if ok {
		newItem.ID = len(list.Items) + 1
		list.Items = append(list.Items, newItem)
		models.Checklists[checklistID] = list
		c.JSON(200, gin.H{
			"message": "Item added",
		})
	} else {
		c.JSON(404, gin.H{
			"message": "Could not add item",
		})
	}
}

// PutItem updates an item in a checklist.
func PutItem(c *gin.Context) {
	checklistID, _ := strconv.Atoi(c.Param("id"))
	itemID, _ := strconv.Atoi(c.Param("itemID"))

	var updatedItem models.ChecklistItem

	if err := c.BindJSON(&updatedItem); err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid request",
		})
		return
	}

	list, ok := models.Checklists[checklistID]

	updated := false
	if ok {
		for i, item := range list.Items {
			if item.ID == itemID {
				list.Items[i].Checked = updatedItem.Checked
				updated = true
			}
		}
	}

	if updated {
		models.Checklists[checklistID] = list
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
