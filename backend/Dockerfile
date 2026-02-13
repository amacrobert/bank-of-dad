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
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8001

CMD ["/bank-of-dad"]
