// Package routehandlers provides the route handlers for the application.
package routehandlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

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

// GetChecklists handles the request to get all checklists.
func GetChecklists(c *gin.Context) {
	userID := getUserID(c)
	service, err := db.NewDynamoDBService()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
	} else {
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

// GetSharedChecklists handles the request to get all shared checklists.
func GetSharedChecklists(c *gin.Context) {
	userID := getUserID(c)

	service, err := db.NewDynamoDBService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
		return
	}

	checklists, err := service.GetSharedChecklists(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error getting shared checklists: " + err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"checklists": checklists,
		})
	}
}

// GetChecklist handles the request to get a single checklist.
func GetChecklist(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	service, err := db.NewDynamoDBService()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
		return
	}

	checklist, checklistErr := service.GetChecklist(userID, id)
	items, itemsErr := service.GetChecklistItems(userID, id)

	if checklistErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Error getting checklist: " + checklistErr.Error(),
		})
	} else if itemsErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Error getting items: " + itemsErr.Error(),
		})
	} else if checklist.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Checklist does not exist",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"checklist": checklist,
			"items":     items,
		})
	}
}

// GetSharedChecklist handles the request to get a shared checklist.
func GetSharedChecklist(c *gin.Context) {
	userID := getUserID(c)
	checklistID := c.Param("id")

	service, err := db.NewDynamoDBService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
		return
	}

	ownerID, err := service.GetChecklistOwner(userID, checklistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error getting checklist owner: " + err.Error(),
		})
		return
	}

	checklist, checklistErr := service.GetChecklist(ownerID, checklistID)
	items, itemsErr := service.GetChecklistItems(ownerID, checklistID)

	if checklistErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Error getting checklist: " + checklistErr.Error(),
		})
	} else if itemsErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Error getting items: " + itemsErr.Error(),
		})
	} else if checklist.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Checklist does not exist",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"checklist": checklist,
			"items":     items,
		})
	}
}

// PutChecklist handles the request to update a checklist.
func PutChecklist(c *gin.Context) {
	userID := getUserID(c)

	service, err := db.NewDynamoDBService()
	var updatedChecklist models.Checklist

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
		return
	}

	if err := c.BindJSON(&updatedChecklist); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request: " + err.Error(),
		})
	} else {
		updatedChecklist.ID = c.Param("id")
		updatedChecklist.UpdatedAt = time.Now().Format(time.RFC3339)
		err := service.UpdateChecklist(userID, updatedChecklist.ID, &updatedChecklist)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error updating checklist: " + err.Error(),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "Checklist updated",
			})
		}
	}
}

// PutSharedChecklist handles the request to update a shared checklist.
func PutSharedChecklist(c *gin.Context) {
	userID := getUserID(c)
	checklistID := c.Param("id")

	service, err := db.NewDynamoDBService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
		return
	}

	ownerID, err := service.GetChecklistOwner(userID, checklistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error getting checklist owner: " + err.Error(),
		})
		return
	}

	var updatedChecklist models.Checklist
	if err := c.BindJSON(&updatedChecklist); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request: " + err.Error(),
		})
	} else {
		updatedChecklist.ID = checklistID
		updatedChecklist.UpdatedAt = time.Now().Format(time.RFC3339)
		err := service.UpdateChecklist(ownerID, updatedChecklist.ID, &updatedChecklist)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error updating checklist: " + err.Error(),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "Checklist updated",
			})
		}
	}
}

// PostChecklist handles the request to create a new checklist.
func PostChecklist(c *gin.Context) {
	userID := getUserID(c)

	service, err := db.NewDynamoDBService()
	var checklist models.Checklist

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
		return
	}

	if err := c.BindJSON(&checklist); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request: " + err.Error(),
		})
	} else if checklist.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Title is required",
		})
	} else {
		checklist.ID = uuid.New().String()
		checklist.Locked = false
		checklist.CreatedAt = time.Now().Format(time.RFC3339)
		checklist.UpdatedAt = checklist.CreatedAt
		err := service.CreateChecklist(userID, &checklist)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error creating checklist: " + err.Error(),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message":   "Checklist created",
				"checklist": checklist,
			})
		}
	}
}

// DeleteChecklist handles the request to delete a checklist.
func DeleteChecklist(c *gin.Context) {
	userID := getUserID(c)
	id := c.Param("id")

	service, err := db.NewDynamoDBService()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
	} else {
		err := service.DeleteChecklist(userID, id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error deleting checklist: " + err.Error(),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "Checklist deleted",
			})
		}
	}
}

// LeaveSharedChecklist handles the request to remove a user from a shared checklist.
func LeaveSharedChecklist(c *gin.Context) {
	userID := getUserID(c)
	checklistID := c.Param("id")

	service, err := db.NewDynamoDBService()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
	} else {
		err := service.RemoveCollaborator(userID, checklistID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error leaving shared checklist: " + err.Error(),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "Left shared checklist",
			})
		}
	}
}

// GetShareCode handles the request to generate a share code for a checklist.
func GetShareCode(c *gin.Context) {
	userID := getUserID(c)
	checklistID := c.Param("id")

	code, err := sharing.GetShareCode(checklistID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error generating share code: " + err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"code": code,
		})
	}
}

// PostUserToSharedChecklist handles the request to add a user to a shared checklist.
func PostUserToSharedChecklist(c *gin.Context) {
	userID := getUserID(c)
	code := c.Param("code")

	token, err := sharing.GetTokenFromShareCode(code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error getting token from share code: " + err.Error(),
		})
		return
	}

	parsedToken, err := sharing.ParseSharingToken(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error parsing token: " + err.Error(),
		})
		return
	}

	if parsedToken.UserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "You can't add yourself to your own checklist",
		})
		return
	}

	service, err := db.NewDynamoDBService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
		return
	}

	err = service.AddCollaborator(parsedToken.UserID, parsedToken.ChecklistID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error adding user to shared checklist: " + err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "User added to shared checklist",
		})
	}
}

// PostItem handles the request to add an item to a checklist.
func PostItem(c *gin.Context) {
	userID := getUserID(c)
	checklistID := c.Param("id")
	var newItem models.ChecklistItem

	service, err := db.NewDynamoDBService()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
	} else if err := c.BindJSON(&newItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request: " + err.Error(),
		})
	} else {
		newItem.ID = uuid.New().String()
		newItem.Checked = false
		newItem.CreatedAt = time.Now().Format(time.RFC3339)
		newItem.UpdatedAt = newItem.CreatedAt

		err := service.CreateChecklistItem(userID, checklistID, &newItem)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error creating item: " + err.Error(),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "Item created",
				"item":    newItem,
			})
		}
	}
}

// PostSharedItem handles the request to add an item to a shared checklist.
func PostSharedItem(c *gin.Context) {
	userID := getUserID(c)
	checklistID := c.Param("id")

	service, err := db.NewDynamoDBService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
		return
	}

	ownerID, err := service.GetChecklistOwner(userID, checklistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error getting checklist owner: " + err.Error(),
		})
		return
	}

	var newItem models.ChecklistItem
	if err := c.BindJSON(&newItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request: " + err.Error(),
		})
	} else {
		newItem.ID = uuid.New().String()
		newItem.Checked = false
		newItem.CreatedAt = time.Now().Format(time.RFC3339)
		newItem.UpdatedAt = newItem.CreatedAt

		err := service.CreateChecklistItem(ownerID, checklistID, &newItem)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error creating item: " + err.Error(),
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "Item created",
				"item":    newItem,
			})
		}
	}
}

// PutItem handles the request to update an item in a checklist.
func PutItem(c *gin.Context) {
	userID := getUserID(c)
	checklistID := c.Param("id")
	itemID := c.Param("itemID")

	var updatedItem models.ChecklistItem

	if err := c.BindJSON(&updatedItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request",
		})
		return
	}

	service, err := db.NewDynamoDBService()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
		return
	}

	updatedItem.UpdatedAt = time.Now().Format(time.RFC3339)
	err = service.UpdateChecklistItem(userID, checklistID, itemID, &updatedItem)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error updating item: " + err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Item updated",
		})
	}
}

// PutSharedItem handles the request to update an item in a shared checklist.
func PutSharedItem(c *gin.Context) {
	userID := getUserID(c)
	checklistID := c.Param("id")
	itemID := c.Param("itemID")

	service, err := db.NewDynamoDBService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
		return
	}

	ownerID, err := service.GetChecklistOwner(userID, checklistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error getting checklist owner: " + err.Error(),
		})
		return
	}

	var updatedItem models.ChecklistItem
	if err := c.BindJSON(&updatedItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request",
		})
		return
	}

	updatedItem.UpdatedAt = time.Now().Format(time.RFC3339)
	err = service.UpdateChecklistItem(ownerID, checklistID, itemID, &updatedItem)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error updating item: " + err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Item updated",
		})
	}
}

// PutAllItems handles the request to update all items in a checklist.
func PutAllItems(c *gin.Context) {
	userID := getUserID(c)
	checklistID := c.Param("id")
	checked := c.Query("checked") == "true"
	service, err := db.NewDynamoDBService()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
		return
	}

	err = service.UpdateChecklistItems(userID, checklistID, checked)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error updating items: " + err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Items updated",
		})
	}
}

// PutAllSharedItems handles the request to update all items in a shared checklist.
func PutAllSharedItems(c *gin.Context) {
	userID := getUserID(c)
	checklistID := c.Param("id")
	checked := c.Query("checked") == "true"

	service, err := db.NewDynamoDBService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
		return
	}

	ownerID, err := service.GetChecklistOwner(userID, checklistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error getting checklist owner: " + err.Error(),
		})
		return
	}

	err = service.UpdateChecklistItems(ownerID, checklistID, checked)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error updating items: " + err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Items updated",
		})
	}
}

// DeleteItem handles the request to delete an item from a checklist.
func DeleteItem(c *gin.Context) {
	userID := getUserID(c)
	checklistID := c.Param("id")
	itemID := c.Param("itemID")

	service, err := db.NewDynamoDBService()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
		return
	}

	err = service.DeleteChecklistItem(userID, checklistID, itemID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error deleting item: " + err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Item deleted",
		})
	}
}

// DeleteSharedItem handles the request to delete an item from a shared checklist.
func DeleteSharedItem(c *gin.Context) {
	userID := getUserID(c)
	checklistID := c.Param("id")
	itemID := c.Param("itemID")

	service, err := db.NewDynamoDBService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
		return
	}

	ownerID, err := service.GetChecklistOwner(userID, checklistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error getting checklist owner: " + err.Error(),
		})
		return
	}

	err = service.DeleteChecklistItem(ownerID, checklistID, itemID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error deleting item: " + err.Error(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"message": "Item deleted",
		})
	}
}

// PostUser handles the request to create a new user.
func PostUser(c *gin.Context) {
	var user models.User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error parsing user info: " + err.Error(),
		})
		return
	}

	service, err := db.NewDynamoDBService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error setting up DynamoDBService: " + err.Error(),
		})
		return
	}

	existingUser, err := service.GetUser(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error checking for user: " + err.Error(),
		})
		return
	}

	if existingUser.ID != "" {
		err = service.UpdateUser(user.ID, user.Email, user.Picture)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error updating user: " + err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "User updated",
		})
	} else {
		err = service.CreateUser(user.ID, user.Email, user.Picture)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error creating user: " + err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "User created",
		})
	}
}
