# Stage 1: Build the Go binary
FROM golang:alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Enable CGO and build the Go app
ENV CGO_ENABLED=1
RUN go build -o main ./cmd/server

# Stage 2: Run the Go binary
FROM alpine

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Expose port 8083 to the outside world
EXPOSE 8083

# Command to run the executable
CMD ["./main"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD wget --no-verbose --spider http://localhost:8083/health || exit 1