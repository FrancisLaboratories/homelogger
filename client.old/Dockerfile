# Stage 1: Build the Next.js app
FROM node:24-alpine AS builder

WORKDIR /app

# Install dependencies
COPY package.json package-lock.json ./
RUN npm install

# Copy the rest of the application code
COPY . .

ARG NEXT_TELEMETRY_DISABLED=1

# Build the Next.js app
RUN npm run build

# Stage 2: Serve the app with Nginx
FROM nginx:stable-alpine

# Copy the built app to the Nginx HTML directory
COPY --from=builder /app/out /usr/share/nginx/html

# Expose port 3000
EXPOSE 3000

# Use the provided nginx.conf file (mounted via Docker Compose)
CMD ["nginx", "-g", "daemon off;"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD wget --no-verbose --spider http://localhost:3000/health || exit 1

