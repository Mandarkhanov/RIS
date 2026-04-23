FROM golang:alpine AS builder

WORKDIR /app
COPY . .

WORKDIR /app/worker-service
RUN go mod download
RUN go build -o app_binary ./cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/worker-service/app_binary .
CMD ["./app_binary"]