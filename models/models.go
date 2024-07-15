// Package models provides the data models for the application.
package models

// Checklist is a collection of ChecklistItems.
type Checklist struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Locked        bool     `json:"locked"`
	Collaborators []string `json:"collaborators"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// ChecklistItem is a single item in a checklist.
type ChecklistItem struct {
	ID        string `json:"id"`
	Content   string `json:"content"`
	Checked   bool   `json:"checked"`
	Order     int    `json:"order"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"update_at"`
}
