FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o litefaas cmd/litefaas/main.go

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/litefaas /usr/local/bin/

EXPOSE 8080

CMD ["litefaas"]
