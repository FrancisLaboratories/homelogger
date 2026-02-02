# Stage 1: Build the Next.js app
FROM node:24-alpine AS client-builder

WORKDIR /client

# Install dependencies
COPY client/package.json client/package-lock.json ./
RUN npm ci --silent

# Copy the rest of the application code
COPY client/ .

ARG NEXT_TELEMETRY_DISABLED=1

# Build the Next.js app (expects build to produce `out` or the static export)
RUN npm run build

# Stage 2: Build the Go binary
FROM golang:1-alpine AS server-builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go.mod from the server directory
COPY server/go.mod ./

# Download dependencies
RUN go mod download

# Copy only the server source into the build context
COPY server/ .

ENV CGO_ENABLED=1
RUN go build -o main ./cmd/server

# Stage 3: Final image based on Alpine, running nginx + the Go server
FROM alpine:latest as final

RUN apk add --no-cache ca-certificates nginx

# Ensure the runtime working directory matches expectations in server code
WORKDIR /root

# Copy the built static site (from client build). If your Next.js build exports static files
# they should be in /client/out; adjust if your build places them elsewhere.
COPY --from=client-builder /client/out /usr/share/nginx/html

# Overwrite nginx main config with a minimal one that includes conf.d/*.conf
COPY nginx-main.conf /etc/nginx/nginx.conf

# Place the server block into nginx's conf.d (renamed)
COPY nginx-default.conf /etc/nginx/conf.d/default.conf

# Copy the Go binary
COPY --from=server-builder /app/main /usr/local/bin/main
RUN chmod +x /usr/local/bin/main

# Copy healthcheck script and use a single unified HEALTHCHECK that calls it.
COPY prod.healthcheck.sh /usr/local/bin/healthcheck.sh
RUN chmod +x /usr/local/bin/healthcheck.sh

# Expose ports (80 for nginx, 8083 for Go service)
EXPOSE 80 8083

# Start both nginx and the Go server 
CMD nginx -g "daemon off;" & /usr/local/bin/main

# The healthcheck script checks the frontend `/health` and backend `/api/health` endpoints
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD "/usr/local/bin/healthcheck.sh http://localhost/health http://localhost/api/health || exit 1"
