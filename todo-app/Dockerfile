FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o todo-app ./cmd

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/config /app/config
COPY --from=builder /app/keys /app/keys
COPY --from=builder /app/todo-app /app/todo-app

EXPOSE 8082
CMD ["/app/todo-app"]