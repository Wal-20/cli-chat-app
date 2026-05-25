# --- Build stage ---
FROM golang:1.24-alpine AS build

WORKDIR /app

RUN apk add --no-cache bash upx tzdata

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG SERVER_URL
ARG JWT_SECRET
ENV SERVER_URL=${SERVER_URL}
ENV JWT_SECRET=${JWT_SECRET}

# Build server and client binaries. --no-source: use the build args above
# instead of sourcing a baked-in .env, so the URL/secret come from compose.
RUN chmod +x ./build.sh && ./build.sh --no-source

# --- Runtime stage ---
FROM alpine:latest

WORKDIR /app

# Copy built binaries and releases
COPY --from=build /app/releases ./releases
COPY --from=build /app/releases/server ./server

EXPOSE 8080

CMD ["./server"]
