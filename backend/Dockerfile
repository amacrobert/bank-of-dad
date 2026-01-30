# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bank-of-dad

# Runtime stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /bank-of-dad /bank-of-dad

# Create data directory for SQLite
RUN mkdir -p /data

ENV DATABASE_PATH=/data/bankodad.db

EXPOSE 8001

VOLUME ["/data"]

CMD ["/bank-of-dad"]
