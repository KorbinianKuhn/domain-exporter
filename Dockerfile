# Build

FROM golang:1.27.3-alpine3.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /build ./cmd/main.go

# Runtime (no shell or package manager, runs as uid 65532)

FROM gcr.io/distroless/static-debian13:nonroot

WORKDIR /domain-exporter
COPY --from=builder /build /app

USER nonroot:nonroot

EXPOSE 2112

ENTRYPOINT ["/app"]
