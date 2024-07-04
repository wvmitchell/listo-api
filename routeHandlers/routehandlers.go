// Package routehandlers provides the route handlers for the application.
package routehandlers

import (
	"checklist-api/db"
	"checklist-api/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"time"
)

// GetChecklists returns all checklists for a user.
func GetChecklists(c *gin.Context) {
	userID := c.GetHeader("userID")

	service, err := db.NewDynamoDBService()

	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
	} else {
		checklists, err := service.GetChecklists(userID)

		if err != nil {
			c.JSON(400, gin.H{
				"message": "Error getting checklists: " + err.Error(),
			})
		} else {
			c.JSON(200, gin.H{
				"checklists": checklists,
			})
		}
	}
}

// GetChecklist returns a single checklist and items.
func GetChecklist(c *gin.Context) {
	userID := c.GetHeader("userID")
	id := c.Param("id")

	fmt.Println("userID: ", userID)
	fmt.Println("id: ", id)

	service, err := db.NewDynamoDBService()

	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
	}

	checklist, checklistErr := service.GetChecklist(userID, id)
	items, itemsErr := service.GetChecklistItems(userID, id)

	if checklistErr != nil {
		c.JSON(400, gin.H{
			"message": "Error getting checklist: " + checklistErr.Error(),
		})
	} else if itemsErr != nil {
		c.JSON(400, gin.H{
			"message": "Error getting items: " + itemsErr.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"checklist": checklist,
			"items":     items,
		})
	}
}

// PostChecklist creates a new checklist.
func PostChecklist(c *gin.Context) {
	service, err := db.NewDynamoDBService()
	userID := c.GetHeader("userID")
	var checklist models.Checklist

	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
	}

	if err := c.BindJSON(&checklist); err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid request: " + err.Error(),
		})
	} else if checklist.Name == "" {
		c.JSON(400, gin.H{
			"message": "Name is required",
		})
	} else {
		checklist.ID = uuid.New().String()
		checklist.Timestamp = time.Now().Format(time.RFC3339)
		err := service.CreateChecklist(userID, &checklist)

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Error creating checklist: " + err.Error(),
			})
		} else {
			c.JSON(200, gin.H{
				"message":   "Checklist created",
				"checklist": checklist,
			})
		}
	}
}

// PostItem adds an item to a checklist.
func PostItem(c *gin.Context) {
	userID := c.GetHeader("userID")
	checklistID := c.Param("id")
	var newItem models.ChecklistItem

	service, err := db.NewDynamoDBService()

	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
	} else if err := c.BindJSON(&newItem); err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid request: " + err.Error(),
		})
	} else {
		newItem.ID = uuid.New().String()
		newItem.Checked = false

		err := service.CreateChecklistItem(userID, checklistID, &newItem)

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Error creating item: " + err.Error(),
			})
		} else {
			c.JSON(200, gin.H{
				"message": "Item created",
				"item":    newItem,
			})
		}
	}
}

// PutItem updates an item in a checklist.
func PutItem(c *gin.Context) {
	userID := c.GetHeader("userID")
	checklistID := c.Param("id")
	itemID := c.Param("itemID")

	var updatedItem models.ChecklistItem

	if err := c.BindJSON(&updatedItem); err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid request",
		})
		return
	}

	service, err := db.NewDynamoDBService()

	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
		return
	}

	err = service.UpdateChecklistItem(userID, checklistID, itemID, &updatedItem)

	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error updating item: " + err.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"message": "Item updated",
		})
	}
}
