FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o backend main.go
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/backend .
EXPOSE 8081
EXPOSE 9/udp
CMD ["./backend"]