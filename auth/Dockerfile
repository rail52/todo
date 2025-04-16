FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o auth-service ./cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/auth-service /app/auth-service

COPY --from=builder /app/keys /app/keys

COPY --from=builder /app/.env /app/.env


EXPOSE 8081

CMD ["/app/auth-service"]
