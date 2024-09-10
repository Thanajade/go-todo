# Build stage
FROM golang:1.20.5-bullseye AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Run tests (no need to specify files, go test will find *_test.go automatically)
RUN go test -v

# Build the application
RUN go build -o go-todo main.go

# Run stage
FROM golang:1.23.1-bullseye

# Set the working directory inside the container
WORKDIR /app

# Copy the executable from the build stage
COPY --from=builder /app/go-todo .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./go-todo"]