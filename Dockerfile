# Use the Go image as the base for building
FROM golang:1.20 AS builder

# Set the working directory in the container
WORKDIR /app

# Copy go.mod and go.sum first to download dependencies
COPY go.mod go.sum ./

# Download dependencies (cached)
RUN go mod download

# Copy the application source code
COPY cmd/ ./cmd/
COPY config/ ./config/
COPY internal/ ./internal/
COPY queries/ ./queries/

# Build the application
WORKDIR /app/cmd
RUN go build -o /app/out

# Create a minimal runtime image
FROM ubuntu:latest

# Set the working directory in the runtime container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/out .

# Expose the port your app uses
EXPOSE 8080

# Run the application
CMD ["./out"]