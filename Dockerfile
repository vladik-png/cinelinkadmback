FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod tidy
COPY . .
RUN go build -o admin-aws

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/admin-aws .
COPY templates ./templates
EXPOSE 8080
CMD ["./admin-aws"]
