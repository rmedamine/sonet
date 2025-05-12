# Stage 1: Build the frontend
FROM node:current-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ .
RUN npm run build

# Stage 2: Build the backend
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app/backend
RUN apk add --no-cache gcc musl-dev
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN go build -o main .

# Stage 3: Final stage
FROM alpine:latest

WORKDIR /app

# Install Node.js and npm
RUN apk add --no-cache nodejs npm

# Copy frontend build
COPY --from=frontend-builder /app/frontend/ ./frontend
# Copy backend binary
COPY --from=backend-builder /app/backend ./backend

# Copy start script
COPY start.sh ./

RUN chmod +x start.sh

EXPOSE 3000 8000

CMD ["./start.sh"] 