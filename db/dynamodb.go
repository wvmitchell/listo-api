// Package db sets up the database connection and provides the query functions for the application.
package db

import (
	"checklist-api/models"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// DynamoDBService is a struct that holds the DynamoDB client.
type DynamoDBService struct {
	Client *dynamodb.Client
}

// NewDynamoDBService creates a new DynamoDBService object.
func NewDynamoDBService() (*DynamoDBService, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))

	if err != nil {
		return nil, fmt.Errorf("failed to load configuration, %v", err)
	}

	svc := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		str := os.Getenv("ENVIRONMENT")
		if str == "development" {
			o.BaseEndpoint = aws.String("http://localhost:8000")
		}
	})

	return &DynamoDBService{
		Client: svc,
	}, nil
}

// EnsureTableExists checks if the table exists and creates it if it does not.
func (d *DynamoDBService) EnsureTableExists(tableName string, createTableFunc func(svc *dynamodb.Client) error) error {
	_, err := d.Client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})

	if err != nil {
		var notFoundErr *types.ResourceNotFoundException
		if ok := errors.As(err, &notFoundErr); ok {
			fmt.Printf("Table %s does not exist, creating table...\n", tableName)
			err := createTableFunc(d.Client)

			if err != nil {
				fmt.Printf("Error creating table %s: %v\n", tableName, err)
			}
		}
		fmt.Printf("Error describing table %s: %v\n", tableName, err)

	}

	fmt.Printf("Table %s already exists\n", tableName)
	return nil
}

// GetChecklists retrieves all checklists for a user.
func (d *DynamoDBService) GetChecklists(userID string) ([]models.Checklist, error) {
	output, err := d.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String("Checklists"),
		KeyConditionExpression: aws.String("PK = :pk"),
		FilterExpression:       aws.String("Entity = :entity"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     &types.AttributeValueMemberS{Value: "USER#" + userID},
			":entity": &types.AttributeValueMemberS{Value: "CHECKLIST"},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query table, %v", err)
	}

	checklists := []models.Checklist{}
	for _, item := range output.Items {
		checklistID := strings.Split(item["SK"].(*types.AttributeValueMemberS).Value, "#")[1]
		checklist, err := d.GetChecklist(userID, checklistID)
		if err != nil {
			return nil, fmt.Errorf("failed to get checklist, %v", err)
		}

		checklists = append(checklists, checklist)
	}

	return checklists, nil
}

// GetSharedChecklists retrieves all checklists shared with a user.
func (d *DynamoDBService) GetSharedChecklists(userID string) ([]models.Checklist, error) {
	output, err := d.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String("ChecklistCollaborators"),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "USER#" + userID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query table, %v", err)
	}

	checklists := []models.Checklist{}
	for _, item := range output.Items {
		checklistID := strings.Split(item["SK"].(*types.AttributeValueMemberS).Value, "#")[1]
		userID := item["OwnerID"].(*types.AttributeValueMemberS).Value
		checklist, err := d.GetChecklist(userID, checklistID)
		if err != nil {
			return nil, fmt.Errorf("failed to get checklist, %v", err)
		}
		checklists = append(checklists, checklist)
	}

	return checklists, nil
}

// GetChecklist retrieves a single checklist.
func (d *DynamoDBService) GetChecklist(userID string, checklistID string) (models.Checklist, error) {
	output, err := d.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String("Checklists"),
		KeyConditionExpression: aws.String("PK = :pk AND SK = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "USER#" + userID},
			":sk": &types.AttributeValueMemberS{Value: "CHECKLIST#" + checklistID},
		},
	})

	if err != nil {
		return models.Checklist{}, fmt.Errorf("failed to query table, %v", err)
	}

	if len(output.Items) == 0 {
		return models.Checklist{}, nil
	}

	collaborators, err := d.GetChecklistCollaborators(userID, checklistID)
	if err != nil {
		return models.Checklist{}, fmt.Errorf("failed to get checklist collaborators, %v", err)
	}

	item := output.Items[0]
	checklist := models.Checklist{
		ID:            strings.Split(item["SK"].(*types.AttributeValueMemberS).Value, "#")[1],
		Title:         item["Title"].(*types.AttributeValueMemberS).Value,
		Locked:        item["Locked"].(*types.AttributeValueMemberBOOL).Value,
		Collaborators: collaborators,
		CreatedAt:     item["CreatedAt"].(*types.AttributeValueMemberS).Value,
		UpdatedAt:     item["UpdatedAt"].(*types.AttributeValueMemberS).Value,
	}

	return checklist, nil
}

// GetChecklistItems retrieves the items for a checklist.
func (d *DynamoDBService) GetChecklistItems(userID string, checklistID string) ([]models.ChecklistItem, error) {
	output, err := d.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String("Checklists"),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		FilterExpression:       aws.String("Entity = :entity"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     &types.AttributeValueMemberS{Value: "USER#" + userID},
			":sk":     &types.AttributeValueMemberS{Value: "CHECKLIST#" + checklistID + "ITEM#"},
			":entity": &types.AttributeValueMemberS{Value: "ITEM"},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query table, %v", err)
	}

	checklistItems := []models.ChecklistItem{}

	for _, item := range output.Items {
		orderingVal, err := strconv.Atoi(item["Ordering"].(*types.AttributeValueMemberN).Value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse order for item")
		}

		checklistItem := models.ChecklistItem{
			ID:        strings.Split(item["SK"].(*types.AttributeValueMemberS).Value, "ITEM#")[1],
			Content:   item["Content"].(*types.AttributeValueMemberS).Value,
			Checked:   item["Checked"].(*types.AttributeValueMemberBOOL).Value,
			Ordering:  orderingVal,
			CreatedAt: item["CreatedAt"].(*types.AttributeValueMemberS).Value,
			UpdatedAt: item["UpdatedAt"].(*types.AttributeValueMemberS).Value,
		}

		checklistItems = append(checklistItems, checklistItem)
	}

	return checklistItems, nil
}

// CreateChecklist creates a new checklist in the database.
func (d *DynamoDBService) CreateChecklist(userID string, checklist *models.Checklist) error {
	_, err := d.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("Checklists"),
		Item: map[string]types.AttributeValue{
			"PK":        &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK":        &types.AttributeValueMemberS{Value: "CHECKLIST#" + checklist.ID},
			"Entity":    &types.AttributeValueMemberS{Value: "CHECKLIST"},
			"Title":     &types.AttributeValueMemberS{Value: checklist.Title},
			"Locked":    &types.AttributeValueMemberBOOL{Value: checklist.Locked},
			"CreatedAt": &types.AttributeValueMemberS{Value: checklist.CreatedAt},
			"UpdatedAt": &types.AttributeValueMemberS{Value: checklist.UpdatedAt},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to put item, %v", err)
	}

	return nil
}

// UpdateChecklist updates a checklist in the database.
func (d *DynamoDBService) UpdateChecklist(userID string, checklistID string, checklist *models.Checklist) error {
	_, err := d.Client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String("Checklists"),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "CHECKLIST#" + checklistID},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":title":     &types.AttributeValueMemberS{Value: checklist.Title},
			":locked":    &types.AttributeValueMemberBOOL{Value: checklist.Locked},
			":updatedAt": &types.AttributeValueMemberS{Value: checklist.UpdatedAt},
		},
		ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
		UpdateExpression:    aws.String("SET Title = :title, Locked = :locked, UpdatedAt = :updatedAt"),
	})

	if err != nil {
		return fmt.Errorf("failed to update item, %v", err)
	}
	return nil
}

// DeleteChecklist deletes a checklist and all associated items from the database, if unlocked.
func (d *DynamoDBService) DeleteChecklist(userID string, checklistID string) error {
	checklist, err := d.GetChecklist(userID, checklistID)
	if err != nil {
		return fmt.Errorf("failed to get checklist, %v", err)
	} else if checklist.Locked {
		return fmt.Errorf("checklist is locked")
	}

	items, err := d.GetChecklistItems(userID, checklistID)
	if err != nil {
		return fmt.Errorf("failed to get checklist items, %v", err)
	}

	if len(items) > 0 {
		var deleteRequests []types.WriteRequest

		for _, item := range items {
			deleteRequests = append(deleteRequests, types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: map[string]types.AttributeValue{
						"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
						"SK": &types.AttributeValueMemberS{Value: "CHECKLIST#" + checklistID + "ITEM#" + item.ID},
					},
				},
			})
		}

		_, err = d.Client.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				"Checklists": deleteRequests,
			},
		})

		if err != nil {
			return fmt.Errorf("failed to delete checklist items, %v", err)
		}
	}

	_, err = d.Client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String("Checklists"),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "CHECKLIST#" + checklistID},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete checklist, %v", err)
	}

	err = d.deleteChecklistCollaborators(userID, checklistID)
	if err != nil {
		return fmt.Errorf("failed to delete checklist collaborators, %v", err)
	}

	return nil
}

// deleteChecklistCollaborators deletes all collaborators for a checklist.
// using the GSI to find all collaborators for a checklist.
// userID is the owner of the checklist.
func (d *DynamoDBService) deleteChecklistCollaborators(userID string, checklistID string) error {
	output, err := d.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String("ChecklistCollaborators"),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :pk AND GSI1SK = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "USER#" + userID},
			":sk": &types.AttributeValueMemberS{Value: "CHECKLIST#" + checklistID},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to query table, %v", err)
	}

	if len(output.Items) > 0 {
		var deleteRequests []types.WriteRequest

		for _, item := range output.Items {
			deleteRequests = append(deleteRequests, types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: map[string]types.AttributeValue{
						"PK": &types.AttributeValueMemberS{Value: item["PK"].(*types.AttributeValueMemberS).Value},
						"SK": &types.AttributeValueMemberS{Value: item["SK"].(*types.AttributeValueMemberS).Value},
					},
				},
			})
		}

		_, err = d.Client.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				"ChecklistCollaborators": deleteRequests,
			},
		})

		if err != nil {
			return fmt.Errorf("failed to delete checklist collaborators, %v", err)
		}
	}

	return nil
}

// AddCollaborator adds a collaborator to a checklist.
func (d *DynamoDBService) AddCollaborator(userID string, checklistID string, collaboratorID string) error {
	_, err := d.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("ChecklistCollaborators"),
		Item: map[string]types.AttributeValue{
			"PK":             &types.AttributeValueMemberS{Value: "USER#" + collaboratorID},
			"SK":             &types.AttributeValueMemberS{Value: "CHECKLIST#" + checklistID},
			"OwnerID":        &types.AttributeValueMemberS{Value: userID},
			"GSI1PK":         &types.AttributeValueMemberS{Value: "USER#" + userID},
			"GSI1SK":         &types.AttributeValueMemberS{Value: "CHECKLIST#" + checklistID},
			"CollaboratorID": &types.AttributeValueMemberS{Value: collaboratorID},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to put item, %v", err)
	}

	return nil
}

// RemoveCollaborator removes a collaborator from a checklist.
func (d *DynamoDBService) RemoveCollaborator(collaboratorID string, checklistID string) error {
	_, err := d.Client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String("ChecklistCollaborators"),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + collaboratorID},
			"SK": &types.AttributeValueMemberS{Value: "CHECKLIST#" + checklistID},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to delete item, %v", err)
	}

	return nil
}

// GetChecklistCollaborators retrieves all collaborators for a checklist
func (d *DynamoDBService) GetChecklistCollaborators(userID string, checklistID string) ([]models.Collaborator, error) {
	output, err := d.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String("ChecklistCollaborators"),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :pk AND GSI1SK = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "USER#" + userID},
			":sk": &types.AttributeValueMemberS{Value: "CHECKLIST#" + checklistID},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to query table, %v", err)
	}

	collaborators := []models.Collaborator{}
	for _, item := range output.Items {
		collaboratorID := strings.Split(item["PK"].(*types.AttributeValueMemberS).Value, "#")[1]
		user, err := d.GetUser(collaboratorID)
		collaborator := models.Collaborator{
			Email:   user.Email,
			Picture: user.Picture,
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get user, %v", err)
		}
		collaborators = append(collaborators, collaborator)
	}
	owner, err := d.GetUser(userID)
	ownerCollaborator := models.Collaborator{
		Email:   owner.Email,
		Picture: owner.Picture,
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user, %v", err)
	}

	collaborators = append(collaborators, ownerCollaborator)

	return collaborators, nil
}

// GetChecklistOwner retrieves the owner of a checklist.
func (d *DynamoDBService) GetChecklistOwner(userID string, checklistID string) (string, error) {
	output, err := d.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String("ChecklistCollaborators"),
		KeyConditionExpression: aws.String("PK = :pk AND SK = :sk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "USER#" + userID},
			":sk": &types.AttributeValueMemberS{Value: "CHECKLIST#" + checklistID},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to query table, %v", err)
	}

	return output.Items[0]["OwnerID"].(*types.AttributeValueMemberS).Value, nil
}

// CreateChecklistItem creates a new item in a checklist.
func (d *DynamoDBService) CreateChecklistItem(userID string, checklistID string, item *models.ChecklistItem) error {
	_, err := d.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("Checklists"),
		Item: map[string]types.AttributeValue{
			"PK":        &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK":        &types.AttributeValueMemberS{Value: "CHECKLIST#" + checklistID + "ITEM#" + item.ID},
			"Entity":    &types.AttributeValueMemberS{Value: "ITEM"},
			"Content":   &types.AttributeValueMemberS{Value: item.Content},
			"Checked":   &types.AttributeValueMemberBOOL{Value: item.Checked},
			"Ordering":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", item.Ordering)},
			"CreatedAt": &types.AttributeValueMemberS{Value: item.CreatedAt},
			"UpdatedAt": &types.AttributeValueMemberS{Value: item.UpdatedAt},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to put item, %v", err)
	}

	return nil
}

// UpdateChecklistItem updates an item in a checklist.
func (d *DynamoDBService) UpdateChecklistItem(userID string, checklistID string, itemID string, item *models.ChecklistItem) error {
	_, err := d.Client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String("Checklists"),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "CHECKLIST#" + checklistID + "ITEM#" + itemID},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":content":   &types.AttributeValueMemberS{Value: item.Content},
			":checked":   &types.AttributeValueMemberBOOL{Value: item.Checked},
			":ordering":  &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", item.Ordering)},
			":updatedAt": &types.AttributeValueMemberS{Value: item.UpdatedAt},
		},
		ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
		UpdateExpression:    aws.String("SET Content = :content, Checked = :checked, Ordering = :ordering, UpdatedAt = :updatedAt"),
		ReturnValues:        types.ReturnValueAllNew,
	})
	if err != nil {
		return fmt.Errorf("failed to update item, %v", err)
	}

	return nil
}

// UpdateChecklistItems updates all items in a checklist.
// currently only supports 100 total items, and is only for checking/unchecking all items.
func (d *DynamoDBService) UpdateChecklistItems(userID string, checklistID string, checked bool) error {
	items, err := d.GetChecklistItems(userID, checklistID)

	if err != nil {
		return fmt.Errorf("failed to get checklist items, %v", err)
	}

	if len(items) == 0 {
		return nil
	}

	transactItems := []types.TransactWriteItem{}

	for _, item := range items {
		item.UpdatedAt = time.Now().Format(time.RFC3339)
		transactItems = append(transactItems, types.TransactWriteItem{
			Update: &types.Update{
				TableName: aws.String("Checklists"),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
					"SK": &types.AttributeValueMemberS{Value: "CHECKLIST#" + checklistID + "ITEM#" + item.ID},
				},
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":checked":   &types.AttributeValueMemberBOOL{Value: checked},
					":updatedAt": &types.AttributeValueMemberS{Value: item.UpdatedAt},
				},
				ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
				UpdateExpression:    aws.String("SET Checked = :checked, UpdatedAt = :updatedAt"),
			},
		})
	}

	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: transactItems,
	}

	_, err = d.Client.TransactWriteItems(context.TODO(), input)

	if err != nil {
		return fmt.Errorf("failed to update items, %v", err)
	}

	return nil
}

// DeleteChecklistItem deletes an item from a checklist.
func (d *DynamoDBService) DeleteChecklistItem(userID string, checklistID string, itemID string) error {
	_, err := d.Client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String("Checklists"),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "CHECKLIST#" + checklistID + "ITEM#" + itemID},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to delete item, %v", err)
	}

	return nil
}

// GetUser retrieves a user from the database.
func (d *DynamoDBService) GetUser(userID string) (models.User, error) {
	response, err := d.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String("Users"),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: userID},
		},
	})

	if err != nil {
		return models.User{}, fmt.Errorf("failed to get item, %v", err)
	}

	user := models.User{}
	err = attributevalue.UnmarshalMap(response.Item, &user)

	if err != nil {
		return models.User{}, fmt.Errorf("failed to unmarshal item, %v", err)
	} else if user.ID == "" {
		return models.User{}, nil
	}

	return user, nil
}

// CreateUser creates a new user in the database.
func (d *DynamoDBService) CreateUser(userID string, email string, picture string) error {
	_, err := d.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("Users"),
		Item: map[string]types.AttributeValue{
			"ID":      &types.AttributeValueMemberS{Value: userID},
			"Email":   &types.AttributeValueMemberS{Value: email},
			"Picture": &types.AttributeValueMemberS{Value: picture},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create user, %v", err)
	}

	err = d.createIntroductoryListoForUser(userID)

	if err != nil {
		return fmt.Errorf("failed to create introductory listo, %v", err)
	}

	return nil
}

// UpdateUser updates a user in the database.
func (d *DynamoDBService) UpdateUser(userID string, email string, picture string) error {
	_, err := d.Client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String("Users"),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: userID},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email":   &types.AttributeValueMemberS{Value: email},
			":picture": &types.AttributeValueMemberS{Value: picture},
		},
		ConditionExpression: aws.String("attribute_exists(ID)"),
		UpdateExpression:    aws.String("SET Email = :email, Picture = :picture"),
	})

	if err != nil {
		return fmt.Errorf("failed to update user, %v", err)
	}

	return nil
}

// createIntroductoryListoForUser creates a new listo for a user with introductory content.
func (d *DynamoDBService) createIntroductoryListoForUser(userID string) error {
	// create new checklist for user
	checklist := models.Checklist{
		ID:        uuid.New().String(),
		Title:     "My First Listo",
		Locked:    false,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}
	// save the checklist in the db
	err := d.CreateChecklist(userID, &checklist)

	if err != nil {
		return fmt.Errorf("failed to create checklist, %v", err)
	}

	firstListContent := []string{
		"Edit the title of this Listo by clicking on the title. Your changes will be saved automatically.",
		"Edit this item by clicking on it, making your changes, and clicking away, or <return>",
		"Add a new item to your Listo next to the + icon",
		"You can make a multi-line item by pressing <shift> + <return>",
		"Mark this item as done, by clicking on the checkbox",
		"Reorder this item by dragging it somewhere else, and dropping it",
		"Delete your checked items by selecting \"Delete Checked\" from the options dropdown",
		"Lock your Listo by selecting \"Lock\" from the options dropdown. You'll still be able to check/uncheck items, but can't change them. This is handy if you have checklists that you need to reuse.",
		"Share your Listo with others by selecting \"Share\" from the options dropdown. You can share with anyone, even if they don't have an account yet.",
		"Have fun!",
	}

	for i, content := range firstListContent {
		item := models.ChecklistItem{
			ID:        uuid.New().String(),
			Content:   content,
			Checked:   false,
			Ordering:  i,
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
		}
		err = d.CreateChecklistItem(userID, checklist.ID, &item)

		if err != nil {
			return fmt.Errorf("failed to create item, %v", err)
		}
	}

	return nil
}
