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

## Deploying to AWS

- The app is deployed to AWS using ECS and ECR. The Dockerfile is included in the repo.
- build the docker image using `docker build -t checklist-api .`
- tag the image using `docker tag checklist-api:latest <aws_account_id>.dkr.ecr.<region>.amazonaws.com/checklist-api:latest`
- push the image to ECR using `docker push <aws_account_id>.dkr.ecr.<region>.amazonaws.com/checklist-api:latest`
- update the task definition in the ECS console to use the new image
- update the service in the ECS console to use the new task definition
