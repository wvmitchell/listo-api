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

// CreateChecklistCollaboratorsTable creates the ChecklistCollaborators table.
func CreateChecklistCollaboratorsTable() error {
	service, _ := db.NewDynamoDBService()
	err := service.EnsureTableExists("ChecklistCollaborators", createChecklistCollaboratorsTableMigration)

	if err != nil {
		fmt.Printf("Error creating table ChecklistCollaborators: %v\n", err)
	}
	return err
}

func createChecklistCollaboratorsTableMigration(svc *dynamodb.Client) error {
	_, err := svc.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
		TableName: aws.String("ChecklistCollaborators"),
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("PK"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("SK"), KeyType: types.KeyTypeRange},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("PK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("SK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("GSI1PK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("GSI1SK"), AttributeType: types.ScalarAttributeTypeS},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(50),
			WriteCapacityUnits: aws.Int64(50),
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("GSI1"),
				KeySchema: []types.KeySchemaElement{
					{AttributeName: aws.String("GSI1PK"), KeyType: types.KeyTypeHash},
					{AttributeName: aws.String("GSI1SK"), KeyType: types.KeyTypeRange},
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(50),
					WriteCapacityUnits: aws.Int64(50),
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("Failed to create table, %v", err)
	}

	fmt.Println("Table ChecklistCollaborators created successfully with GSI1")
	return nil
}
