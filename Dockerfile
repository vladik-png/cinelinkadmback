FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o master-server main.go
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/master-server .
EXPOSE 8082
CMD ["./master-server"]