# Listo API

## Description
This is the api for the listo app, written in Go, using the Gin Web Framework. 

## Routes

- `GET /` - Get the status of the app (used for health checks)
- `GET /checklists` - Get all checklists
- `GET /checklists/:id` - Get a single checklist
- `PUT /checklists/:id` - Update a checklist
- `POST /checklist` - Create a new Checklist
- `DELETE /checklist/:id` - Delete a checklist
- `POST /checklists/:id/items` - Create a new item for a checklist
- `PUT /checklists/:id/items/:itemId` - Update an item in a checklist
- `PUT /checklists/:id/items` - Update all items in a Checklist
- `DELETE /checklists/:id/items/:itemId` - Delete an item in a Checklist
- `GET /user/:id` - Get a User
- `POST /user` - Create a new User (and intro checklist)

## Running the app
- Containerize the app using Docker, and the dev environment:
- `docker build --build-arg ENV=dev -t listo_api .`
- Run the container:
- `docker run -p 8080:8080 listo_api`
- Optionally, you can name the container, as well as run it in the background:
- `docker run -d --name listo_api_container -p 8080:8080 listo_api`
- The app should now be running on `http://localhost:8080`
- If necessary, check the logs using `docker logs listo_api_container`

## Deploying to AWS

- The app is deployed to AWS using ECS and ECR. The Dockerfile is included in the repo.
- build the docker image using `docker build --build-arg ENV=prod -t listo_api .`
- tag the image using `docker tag listo_api:latest <aws_account_id>.dkr.ecr.<region>.amazonaws.com/listo_api:latest`
- push the image to ECR using `docker push <aws_account_id>.dkr.ecr.<region>.amazonaws.com/listo_api:latest`
- update the task definition in the ECS console to use the new image
- update the service in the ECS console to use the new task definition
