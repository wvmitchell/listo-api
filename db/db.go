// Package db sets up the database connection and provides the query functions for the application.
package db

import (
	"checklist-api/models"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
		// TODO: Set the endpoint based on the environment
		o.BaseEndpoint = aws.String("http://localhost:8000")
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

	var checklists []models.Checklist
	for _, item := range output.Items {
		collaborators := []string{}

		for _, c := range item["Collaborators"].(*types.AttributeValueMemberL).Value {
			collaborators = append(collaborators, c.(*types.AttributeValueMemberS).Value)
		}

		checklist := models.Checklist{
			ID:            strings.Split(item["SK"].(*types.AttributeValueMemberS).Value, "#")[1],
			Title:         item["Title"].(*types.AttributeValueMemberS).Value,
			Collaborators: collaborators,
			CreatedAt:     item["CreatedAt"].(*types.AttributeValueMemberS).Value,
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

	item := output.Items[0]
	collaborators := []string{}
	for _, c := range item["Collaborators"].(*types.AttributeValueMemberL).Value {
		collaborators = append(collaborators, c.(*types.AttributeValueMemberS).Value)
	}

	checklist := models.Checklist{
		ID:            strings.Split(item["SK"].(*types.AttributeValueMemberS).Value, "#")[1],
		Title:         item["Title"].(*types.AttributeValueMemberS).Value,
		Collaborators: collaborators,
		CreatedAt:     item["CreatedAt"].(*types.AttributeValueMemberS).Value,
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
		checklistItem := models.ChecklistItem{
			ID:        strings.Split(item["SK"].(*types.AttributeValueMemberS).Value, "ITEM#")[1],
			Content:   item["Content"].(*types.AttributeValueMemberS).Value,
			Checked:   item["Checked"].(*types.AttributeValueMemberBOOL).Value,
			CreatedAt: item["CreatedAt"].(*types.AttributeValueMemberS).Value,
			UpdatedAt: item["UpdatedAt"].(*types.AttributeValueMemberS).Value,
		}

		checklistItems = append(checklistItems, checklistItem)
	}

	return checklistItems, nil
}

// CreateChecklist creates a new checklist in the database.
func (d *DynamoDBService) CreateChecklist(userID string, checklist *models.Checklist) error {
	var collaborators []types.AttributeValue
	for _, c := range checklist.Collaborators {
		collaborators = append(collaborators, &types.AttributeValueMemberS{Value: c})
	}

	_, err := d.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("Checklists"),
		Item: map[string]types.AttributeValue{
			"PK":            &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK":            &types.AttributeValueMemberS{Value: "CHECKLIST#" + checklist.ID},
			"Entity":        &types.AttributeValueMemberS{Value: "CHECKLIST"},
			"Title":         &types.AttributeValueMemberS{Value: checklist.Title},
			"Collaborators": &types.AttributeValueMemberL{Value: collaborators},
			"CreatedAt":     &types.AttributeValueMemberS{Value: checklist.CreatedAt},
			"UpdatedAt":     &types.AttributeValueMemberS{Value: checklist.UpdatedAt},
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
			":updatedAt": &types.AttributeValueMemberS{Value: checklist.UpdatedAt},
		},
		ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
		UpdateExpression:    aws.String("SET Title = :title, UpdatedAt = :updatedAt"),
	})

	if err != nil {
		return fmt.Errorf("failed to update item, %v", err)
	}
	return nil
}

// DeleteChecklist deletes a checklist and all associated items from the database.
func (d *DynamoDBService) DeleteChecklist(userID string, checklistID string) error {
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

	return nil
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
			":updatedAt": &types.AttributeValueMemberS{Value: item.UpdatedAt},
		},
		ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
		UpdateExpression:    aws.String("SET Content = :content, Checked = :checked, UpdatedAt = :updatedAt"),
		ReturnValues:        types.ReturnValueAllNew,
	})

	if err != nil {
		return fmt.Errorf("failed to update item, %v", err)
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
