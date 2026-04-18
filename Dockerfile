FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /wishlist cmd/api/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /wishlist .
COPY --from=builder /app/internal/web ./internal/web
EXPOSE 8080
CMD ["./wishlist"]