FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o api-gateway ./cmd

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/config /app/config
COPY --from=builder /app/api-gateway /app/api-gateway

EXPOSE 8080
CMD ["/app/api-gateway"]