# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/main ./cmd/main

# Final stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root
COPY --from=builder /out/main ./main
EXPOSE 8000
CMD ["./main"]
