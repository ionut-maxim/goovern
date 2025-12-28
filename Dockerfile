# Build stage
FROM golang:1.25.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o goovernd ./cmd/goovernd

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS connections and terminal info
RUN apk --no-cache add ca-certificates ncurses-terminfo-base

# Set terminal environment for proper rendering
ENV TERM=xterm-256color

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/goovernd .

# Expose SSH server port
EXPOSE 42069

# Run the application
CMD ["./goovernd"]
