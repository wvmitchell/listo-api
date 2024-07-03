// Package migrations provides the functions to create/update the database schema.
package migrations

import (
	"checklist-api/db"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// CreateUsersTable creates the Users table.
func CreateUsersTable() error {
	service, _ := db.NewDynamoDBService()
	err := service.EnsureTableExists("Users", createUsersTableMigration)

	if err != nil {
		fmt.Printf("Error creating table Users: %v\n", err)
	}

	return err
}

func createUsersTableMigration(svc *dynamodb.Client) error {
	_, err := svc.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		TableName: aws.String("Users"),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("ID"), KeyType: types.KeyTypeHash},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("ID"), AttributeType: types.ScalarAttributeTypeS},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	})

	if err != nil {
		return fmt.Errorf("Failed to create table, %v", err)
	}

	fmt.Println("Table Users created successfully")
	return nil
}
