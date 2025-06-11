FROM golang:1.24.4-alpine3.21 AS builder

WORKDIR /app
COPY . .

RUN go mod download

CMD ["go", "run", "cmd/api/main.go"]
