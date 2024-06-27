# Checklist API

## Description
This is the api for the checklist app, written in Go, using the Gin Web Framework. 
Currently under local development, no authorization in place.

## Routes

- `GET /checklists` - Get all checklists
- `GET /checklists/:id` - Get a single checklist
- `POST /checklists` - Create a new Checklist
- `POST /checklists/:id/items` - Create a new item for a checklist
- `PUT /checklists/:id/items/:itemId` - Update an item in a checklist
