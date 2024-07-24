# Use the official Golang image
FROM golang:1.22

# Set work directory
WORKDIR /app

# Copy the Go Modules files
COPY go.mod go.sum ./

# Download and verify Go modules
RUN go mod download && go mod verify

# Copy the source code
COPY . .

# Add build argument for environment
ARG ENV=dev

# Copy the appropriate environment file based on the build argument
RUN if [ "$ENV" = "prod" ] ; then \
        cp .env.prod .env ; \
    else \
        cp .env.dev .env ; \
    fi

# Build the application
RUN go build -o /myapp

# Expose port
EXPOSE 8080

# Start the application
CMD [ "/myapp" ]
