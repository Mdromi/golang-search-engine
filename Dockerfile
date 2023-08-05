# Stage 1: Build the Go binary
FROM golang:1.16 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy only the go.mod and go.sum files first to leverage Docker cache
COPY go.mod go.sum ./

# Download and cache Go dependencies
RUN go mod download

# Copy the rest of the application files
COPY . .

# Build the Go application
RUN go build -o main .

# Stage 2: Create the final minimal image
FROM scratch

# Copy the binary from the builder stage to the final stage
COPY --from=builder /app/main /main

# Set the command to run the application
CMD ["/main"]
