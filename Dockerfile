# Use a lightweight base image for Go
FROM golang:1.23-alpine

# Install necessary tools (e.g., git)
RUN apk add --no-cache git

# Set the working directory inside the container
WORKDIR /app

# Copy Go module files first (for caching dependencies)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application files
COPY . .

# Build the Go binary
RUN go build -o /app/start

# Expose the port your Go Gin app listens on
EXPOSE 8080

# Define the command to run your application
CMD ["/app/start"]