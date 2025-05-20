
# 1. Use official Go base image
FROM golang:1.24-alpine AS build

# 2. Set working directory (think of it as a cd command)
WORKDIR /app

# 3. Copy files into container
COPY . .

# 4. Build the Go binary
RUN go build -o app .

# --- Production Image ---
FROM alpine:latest

# 5. Copy binary from builder
COPY --from=build /app/app /app/app

# 6. Set working directory
WORKDIR /app

# 7. Set environment variable (optional)
# ENV JWT_SECRET=supersecret

# 8. Run the app
CMD ["./app"]
