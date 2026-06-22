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
# Expose the single unified backend port
EXPOSE 8081
# Expose HTTP and HTTPS for Let's Encrypt (if DOMAIN is set)
EXPOSE 80
EXPOSE 443
# Expose UDP port for Wake-on-LAN functionality (optional)
EXPOSE 9/udp
CMD ["./backend"]