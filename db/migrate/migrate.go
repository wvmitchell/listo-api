// Package migrate provides the functions to create/update the database schema.
package migrate

import (
	"checklist-api/db/migrate/migrations"
	"fmt"
)

type migration struct {
	Version   int
	Name      string
	ApplyFunc func() error
}

var migrationList = []migration{
	{1, "1_create_users_table", migrations.CreateUsersTable},
	{2, "2_create_checklists_table", migrations.CreateChecklistsTable},
	// Add new migrations here
}

// RunMigrations runs all the migrations.
func RunMigrations() error {
	for _, migration := range migrationList {
		err := migration.ApplyFunc()
		if err != nil {
			return fmt.Errorf("error applying migration %s: %v", migration.Name, err)
		}
		fmt.Printf("Migration %s applied successfully\n", migration.Name)
	}

	return nil
}
