// Package main	is the entry point of the application.
package main

// ChecklistItem is a single item in a checklist.
type ChecklistItem struct {
	ID      int    `json:"id"`
	Text    string `json:"text"`
	Checked bool   `json:"checked"`
}

// Checklist is a collection of ChecklistItems.
type Checklist struct {
	ID     int             `json:"id"`
	UserID int             `json:"userID"`
	Name   string          `json:"name"`
	Items  []ChecklistItem `json:"items"`
}

var checklists = map[int]Checklist{
	1: {
		ID:     1,
		UserID: 1,
		Name:   "Grocery Shopping",
		Items: []ChecklistItem{
			{ID: 1, Text: "Buy milk", Checked: false},
			{ID: 2, Text: "Buy bread", Checked: true},
			{ID: 3, Text: "Buy eggs", Checked: false},
		},
	},
	2: {
		ID:     2,
		UserID: 2,
		Name:   "Work Tasks",
		Items: []ChecklistItem{
			{ID: 4, Text: "Complete report", Checked: true},
			{ID: 5, Text: "Email client", Checked: false},
			{ID: 6, Text: "Update website", Checked: true},
		},
	},
	3: {
		ID:     3,
		UserID: 1,
		Name:   "Daily Routine",
		Items: []ChecklistItem{
			{ID: 7, Text: "Exercise", Checked: false},
			{ID: 8, Text: "Read a book", Checked: true},
			{ID: 9, Text: "Meditate", Checked: false},
		},
	},
	4: {
		ID:     4,
		UserID: 3,
		Name:   "Weekend Plans",
		Items: []ChecklistItem{
			{ID: 10, Text: "Visit parents", Checked: false},
			{ID: 11, Text: "Go hiking", Checked: true},
			{ID: 12, Text: "Watch a movie", Checked: false},
		},
	},
	5: {
		ID:     5,
		UserID: 2,
		Name:   "Project A",
		Items: []ChecklistItem{
			{ID: 13, Text: "Design UI", Checked: false},
			{ID: 14, Text: "Write code", Checked: true},
			{ID: 15, Text: "Test features", Checked: false},
		},
	},
}
