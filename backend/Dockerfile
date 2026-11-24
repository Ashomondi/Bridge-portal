# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server \
    && CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker

# Worker runtime stage
FROM alpine:3.20 AS worker

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -H -u 10001 appuser

WORKDIR /app
COPY --from=builder /worker /app/worker
USER appuser

CMD ["/app/worker"]

# API server runtime stage (default target)
FROM alpine:3.20 AS server

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -H -u 10001 appuser

WORKDIR /app
COPY --from=builder /server /app/server
USER appuser

EXPOSE 8080
CMD ["/app/server"]
