FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG APP_CMD
RUN go build -o app_binary ./cmd/manager

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app_binary .
CMD ["./app_binary"]