# =========================
# 1) Builder image
# =========================
FROM golang:1.25-alpine AS build

WORKDIR /app

# Install tools needed for build (git, CA certs, tzdata optional)
RUN apk add --no-cache ca-certificates tzdata git

# Copy go.mod & go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the API binary (adjust path if your main is elsewhere)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/api ./cmd/api

# =========================
# 2) Minimal runtime image
# =========================
FROM alpine:3.20

WORKDIR /app

# CA certs needed for HTTPS (Neon, upstream boorus, etc.)
RUN apk add --no-cache ca-certificates tzdata

# Copy compiled binary from builder
COPY --from=build /app/bin/api /app/api

# Set Gin to release mode in container
ENV GIN_MODE=release

# Expose the port your Gin app listens on (assumed 8080)
EXPOSE 80

# Environment variables you MUST provide at runtime:
# - DATABASE_URL=postgres://user:pass@host:port/dbname?sslmode=require
#
# Optionally you can add:
# ENV DATABASE_URL=...

# Run the binary
CMD ["./api"]
