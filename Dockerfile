FROM mirror.gcr.io/library/golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o litefaas cmd/litefaas/main.go

FROM mirror.gcr.io/library/alpine:latest

RUN apk add --no-cache ca-certificates containerd sqlite

WORKDIR /root/

COPY --from=builder /app/litefaas .

EXPOSE 8080

CMD ["./litefaas"]
