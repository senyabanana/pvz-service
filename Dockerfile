FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o pvz-service ./cmd/service/main.go

FROM alpine

WORKDIR /app

COPY --from=builder /app/pvz-service .
COPY .env .env

EXPOSE 8080

CMD ["./pvz-service"]