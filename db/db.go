// Package db sets up the database connection and provides the query functions for the application.
package db

import (
	"checklist-api/models"
	"context"
	"errors"
	"fmt"

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
		o.BaseEndpoint = aws.String("http://localhost:8000") // Use local DynamoDB instance for now, later set based on environment
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
			"Name":          &types.AttributeValueMemberS{Value: checklist.Name},
			"Collaborators": &types.AttributeValueMemberL{Value: collaborators},
			"Timestamp":     &types.AttributeValueMemberS{Value: checklist.Timestamp},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to put item, %v", err)
	}

	return nil
}
