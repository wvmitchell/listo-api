// Package routehandlers provides the route handlers for the application.
package routehandlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"

	"checklist-api/db"
	"checklist-api/models"
	"checklist-api/sharing"
)

func getUserID(c *gin.Context) string {
	sub, exist := c.Get("sub")
	if !exist {
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	return sub.(string)
}

// GetChecklists returns all checklists for a user.
func GetChecklists(c *gin.Context) {
	userID := getUserID(c)
	service, err := db.NewDynamoDBService()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
	} else {

		// Create user if they don't exist, which will also create their first checklist
		user, err := service.GetUser(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error getting user: " + err.Error(),
			})
		} else if user.ID == "" {
			err = service.CreateUser(userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Error creating user: " + err.Error(),
				})
			}
		}

		// Get checklists for a user
		checklists, err := service.GetChecklists(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error getting checklists: " + err.Error(),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"checklists": checklists,
			})
		}
	}
}

// GetSharedChecklists returns all checklists shared with the user (userID) by other users.
func GetSharedChecklists(c *gin.Context) {
	userID := getUserID(c)

	service, err := db.NewDynamoDBService()
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
	}

	checklists, err := service.GetSharedChecklists(userID)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error getting shared checklists: " + err.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"shared checklists": checklists,
		})
	}
}

// GetChecklist returns a single checklist and items.
func GetChecklist(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

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
	} else if checklist.ID == "" {
		c.JSON(404, gin.H{
			"message": "Checklist does not exist",
		})
	} else {
		c.JSON(200, gin.H{
			"checklist": checklist,
			"items":     items,
		})
	}
}

// PutChecklist updates a checklist.
func PutChecklist(c *gin.Context) {
	userID := getUserID(c)

	service, err := db.NewDynamoDBService()
	var updatedChecklist models.Checklist

	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
	}

	if err := c.BindJSON(&updatedChecklist); err != nil {
		c.JSON(400, gin.H{
			"message": "Invalid request: " + err.Error(),
		})
	} else {
		updatedChecklist.ID = c.Param("id")
		updatedChecklist.UpdatedAt = time.Now().Format(time.RFC3339)
		err := service.UpdateChecklist(userID, updatedChecklist.ID, &updatedChecklist)

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Error updating checklist: " + err.Error(),
			})
		} else {
			c.JSON(200, gin.H{
				"message": "Checklist updated",
			})
		}
	}
}

// PostChecklist creates a new checklist.
func PostChecklist(c *gin.Context) {
	userID := getUserID(c)

	service, err := db.NewDynamoDBService()
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
	} else if checklist.Title == "" {
		c.JSON(400, gin.H{
			"message": "Title is required",
		})
	} else {
		checklist.ID = uuid.New().String()
		checklist.Locked = false
		checklist.CreatedAt = time.Now().Format(time.RFC3339)
		checklist.UpdatedAt = checklist.CreatedAt
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

// DeleteChecklist deletes a checklist.
func DeleteChecklist(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	service, err := db.NewDynamoDBService()

	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
	} else {
		err := service.DeleteChecklist(userID, id)

		if err != nil {
			c.JSON(500, gin.H{
				"message": "Error deleting checklist: " + err.Error(),
			})
		} else {
			c.JSON(200, gin.H{
				"message": "Checklist deleted",
			})
		}
	}
}

// GetShareCode returns a share code for a checklist.
func GetShareCode(c *gin.Context) {
	userID := getUserID(c)
	checklistID := c.Param("id")

	code, err := sharing.GetShareCode(checklistID, userID)

	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error generating share code: " + err.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"code": code,
		})
	}
}

func PostUserToSharedChecklist(c *gin.Context) {
	userID := getUserID(c)
	code := c.Param("code")

	token, err := sharing.GetTokenFromShareCode(code)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error getting token from share code: " + err.Error(),
		})
	}

	parsedToken, err := sharing.ParseSharingToken(token)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error parsing token: " + err.Error(),
		})
	}

	if parsedToken.UserID == userID {
		c.JSON(400, gin.H{
			"message": "You can't add yourself to your own checklist",
		})
	}

	service, err := db.NewDynamoDBService()
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
	}

	err = service.AddCollaborator(parsedToken.UserID, parsedToken.ChecklistID, userID)
	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error adding user to shared checklist: " + err.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"message": "User added to shared checklist",
		})
	}
}

// PostItem adds an item to a checklist.
func PostItem(c *gin.Context) {
	userID := getUserID(c)
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
		newItem.CreatedAt = time.Now().Format(time.RFC3339)
		newItem.UpdatedAt = newItem.CreatedAt

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
	userID := getUserID(c)
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

	updatedItem.UpdatedAt = time.Now().Format(time.RFC3339)
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

// PutAllItems updates all items in a checklist. Currently just sets all items to checked/unchecked.
func PutAllItems(c *gin.Context) {
	userID := getUserID(c)
	checklistID := c.Param("id")
	checked := c.Query("checked") == "true"
	service, err := db.NewDynamoDBService()

	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
	}

	err = service.UpdateChecklistItems(userID, checklistID, checked)

	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error updating items: " + err.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"message": "Items updated",
		})
	}
}

// DeleteItem deletes an item from a checklist.
func DeleteItem(c *gin.Context) {
	userID := getUserID(c)
	checklistID := c.Param("id")
	itemID := c.Param("itemID")

	service, err := db.NewDynamoDBService()

	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
		return
	}

	err = service.DeleteChecklistItem(userID, checklistID, itemID)

	if err != nil {
		c.JSON(500, gin.H{
			"message": "Error deleting item: " + err.Error(),
		})
	} else {
		c.JSON(200, gin.H{
			"message": "Item deleted",
		})
	}
}
