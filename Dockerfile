# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates postgresql-client

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy migrations
COPY migrations ./migrations

EXPOSE 8080

CMD ["./main"]
