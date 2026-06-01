FROM golang:1.26.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Собираем goose и приложение
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
COPY . .
RUN go build -o main ./cmd/app/

FROM alpine:latest

WORKDIR /app

# Копируем бинарники и миграции
COPY --from=builder /app/main .
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY migrations ./migrations

EXPOSE 8090

CMD ["./main"]