# Dockerfile
FROM golang:1.19-alpine as builder
WORKDIR /app
COPY . .
RUN go build -o auth-service main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/auth-service .
EXPOSE 8080
CMD ["./auth-service"]
