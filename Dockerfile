# ---- Build stage ----
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/chat-api .

# ---- Runtime stage ----
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the binary
COPY --from=builder /app/chat-api .

# Create uploads directory
RUN mkdir -p /app/uploads

EXPOSE 8080

CMD ["./chat-api"]
