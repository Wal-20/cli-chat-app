# --- Build stage ---
FROM golang:1.23-alpine AS build

WORKDIR /app

RUN apk add --no-cache bash

# Copy dependency files first (for Docker layer caching)
COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN chmod +x build.sh install.sh

ARG SERVER_URL
ENV SERVER_URL=${SERVER_URL}

# Build server and client binaries, no-source depends on server secrets, or args passed into build, instead of .env file in container
RUN ./build.sh --no-source

# --- Runtime stage ---
FROM alpine:latest

WORKDIR /app

# Copy built binaries, install script, and releases
COPY --from=build /app/releases ./releases
COPY --from=build /app/install.sh .
COPY --from=build /app/releases/server ./server

# Expose port for HTTP
EXPOSE 8080

# Start the Go server
CMD ["./server"]

