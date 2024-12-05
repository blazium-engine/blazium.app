# Stage 1: Build the Go binary
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files first
COPY go.mod go.sum ./

# Download and cache Go modules
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o server .

# Stage 2: Create a smaller image to run the binary
FROM alpine:3.18

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/server .

# Copy the static and data folders into the container
COPY ./static ./static
COPY ./data ./data
COPY ./templates ./templates

# Expose the port that the app will run on
EXPOSE 8080

# Run the Go application
CMD ["./server"]
