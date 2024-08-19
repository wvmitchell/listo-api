// Package models provides the data models for the application.
package models

// Checklist is a collection of ChecklistItems.
type Checklist struct {
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	Locked        bool           `json:"locked"`
	Collaborators []Collaborator `json:"collaborators"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
}

// ChecklistItem is a single item in a checklist.
type ChecklistItem struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Checked   bool   `json:"checked"`
	Ordering  int    `json:"ordering"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// User is a user of the application. Most info is actually stored in Auth0.
type User struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

// Collaborator is a user that has been invited to collaborate on a checklist. The ID is removed for extra security, and becuase it's not needed on the client.
type Collaborator struct {
	Email   string `json:"email"`
	Picture string `json:"picture"`
}
