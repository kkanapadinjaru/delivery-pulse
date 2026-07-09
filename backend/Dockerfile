# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install corporate CA cert for TLS inspection proxy (Zscaler)
COPY zscaler-root-ca.crt /usr/local/share/ca-certificates/zscaler-root-ca.crt
RUN apk --no-cache add ca-certificates && update-ca-certificates \
    && cat /usr/local/share/ca-certificates/zscaler-root-ca.crt >> /etc/ssl/certs/ca-certificates.crt
ENV SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/server

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/server .

EXPOSE 8090

ENTRYPOINT ["./server"]
