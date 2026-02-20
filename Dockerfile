# Build stage
FROM golang:alpine AS builder

WORKDIR /app

# Install dependencies required for the build (if any, like CGO)
# RUN apk add --no-cache gcc musl-dev

# Copy modules manifests
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build statically linked binaries
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o main ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o seeder ./cmd/seeder
# Final stage
FROM alpine:3.19

WORKDIR /app

# Add a non-root user for security
RUN adduser -S -D -H -h /app appuser
USER appuser

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/main .
COPY --from=builder /app/seeder .

EXPOSE 8081

CMD ["./main"]
