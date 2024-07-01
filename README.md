# Checklist API

## Description
This is the api for the checklist app, written in Go, using the Gin Web Framework. 

## Routes

- `GET /checklists` - Get all checklists
- `GET /checklists/:id` - Get a single checklist
- `POST /checklists` - Create a new Checklist
- `POST /checklists/:id/items` - Create a new item for a checklist
- `PUT /checklists/:id/items/:itemId` - Update an item in a checklist

## Running the app
- Clone the repo
- Run `go run main.go` in the root directory
- The app will be running on `localhost:8080`

## Notes
- The app is currently under development. Authorization and a database will be added soon.
