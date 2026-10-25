# Build

FROM golang:1.27.3-alpine3.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /build ./cmd/main.go

# Runtime

FROM alpine:3.24

WORKDIR /domain-exporter
COPY --from=builder /build /app

EXPOSE 2112

ENTRYPOINT ["/app"]
