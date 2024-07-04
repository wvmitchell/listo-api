// Package models provides the data models for the application.
package models

// Checklist is a collection of ChecklistItems.
type Checklist struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Collaborators []string `json:"collaborators"`
	CreatedAt     string   `json:"created_at"`
}

// ChecklistItem is a single item in a checklist.
type ChecklistItem struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Checked   bool   `json:"checked"`
	CreatedAt string `json:"created_at"`
	UpdatedAt  string `json:"update_at"`
}
