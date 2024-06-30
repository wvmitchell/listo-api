// Package main	is the entry point of the application.
package main

type item struct {
	Text    string `json:"text"`
	Checked bool   `json:"checked"`
}

type checklist struct {
	ID     int    `json:"id"`
	UserID int    `json:"userID"`
	Name   string `json:"name"`
	Items  []item `json:"items"`
}

var checklists = map[int]checklist{
	1: {
		ID:     1,
		UserID: 1,
		Name:   "Grocery Shopping",
		Items: []item{
			{Text: "Buy milk", Checked: false},
			{Text: "Buy bread", Checked: true},
			{Text: "Buy eggs", Checked: false},
		},
	},
	2: {
		ID:     2,
		UserID: 2,
		Name:   "Work Tasks",
		Items: []item{
			{Text: "Complete report", Checked: true},
			{Text: "Email client", Checked: false},
			{Text: "Update website", Checked: true},
		},
	},
	3: {
		ID:     3,
		UserID: 1,
		Name:   "Daily Routine",
		Items: []item{
			{Text: "Exercise", Checked: false},
			{Text: "Read a book", Checked: true},
			{Text: "Meditate", Checked: false},
		},
	},
	4: {
		ID:     4,
		UserID: 3,
		Name:   "Weekend Plans",
		Items: []item{
			{Text: "Visit parents", Checked: false},
			{Text: "Go hiking", Checked: true},
			{Text: "Watch a movie", Checked: false},
		},
	},
	5: {
		ID:     5,
		UserID: 2,
		Name:   "Project A",
		Items: []item{
			{Text: "Design UI", Checked: false},
			{Text: "Write code", Checked: true},
			{Text: "Test features", Checked: false},
		},
	},
}
