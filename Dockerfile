FROM golang:1.24.5-alpine3.22 AS builder

WORKDIR /app
COPY . .

RUN go mod download

CMD ["go", "run", "cmd/api/main.go"]
