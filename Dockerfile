FROM golang:1.27-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download

CMD ["go", "run", "cmd/api/main.go"]
