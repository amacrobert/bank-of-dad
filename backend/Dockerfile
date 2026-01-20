# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /bank-of-dad

# Runtime stage
FROM alpine:latest

WORKDIR /

COPY --from=builder /bank-of-dad /bank-of-dad

EXPOSE 8001

CMD ["/bank-of-dad"]
