# Stage 1: Build the Go binary
FROM golang:1-alpine AS builder


# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./
COPY go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

RUN go build -o main ./cmd/server

# Stage 2: Run the Go binary
FROM alpine

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Copy the demo data file
COPY --from=builder /app/internal/demo/sample_data.json ./sample_data.json

EXPOSE 3005

# Command to run the executable
CMD ["./main"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD wget --no-verbose --spider http://localhost:3005/api/health || exit 1