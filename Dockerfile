# Stage 1: Build the Next.js app
FROM node:24-alpine AS client-builder

WORKDIR /client

# Install dependencies
COPY client/package.json client/package-lock.json ./
RUN npm ci

# Copy the rest of the application code
COPY client/ .

# Build the Vite app
RUN npm run build

# Stage 2: Build the Go binary
FROM golang:1-alpine AS server-builder


WORKDIR /app

# Copy go.mod from the server directory
COPY server/go.mod ./
COPY server/go.sum ./

# Download dependencies
RUN go mod download

# Copy only the server source into the build context
COPY server/ .

RUN go build -o main ./cmd/server

# Stage 3: Final image — Go binary serves both API + static files
FROM alpine:latest AS final

RUN apk add --no-cache ca-certificates bash curl

# Ensure the runtime working directory matches expectations in server code
WORKDIR /root

# Copy the built static site (from client build).
COPY --from=client-builder /client/dist ./static

# Copy the Go binary
COPY --from=server-builder /app/main /usr/local/bin/main
RUN chmod +x /usr/local/bin/main

# Copy the demo data file
COPY --from=server-builder /app/internal/demo/sample_data.json ./sample_data.json

# Copy healthcheck script
COPY prod.healthcheck.sh /usr/local/bin/healthcheck.sh
RUN chmod +x /usr/local/bin/healthcheck.sh

# Expose the single port the Go server listens on
EXPOSE 3005

# Start the Go server (serves API + static SPA)
CMD ["/usr/local/bin/main"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD ["/usr/local/bin/healthcheck.sh"]
