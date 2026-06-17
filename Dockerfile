# ---------- build stage ----------
FROM golang:1.26-alpine3.24 AS builder

WORKDIR /app

RUN apk --no-cache add bash git make gcc gettext musl-dev

RUN go version
# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build a static binary (CGO off so it runs on a minimal image)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/main ./cmd/main/main.go

# ---------- runtime stage ----------
FROM alpine:3.24

WORKDIR /app

# CA certs for TLS (SMTP/gRPC), tzdata for timezones
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary and config files needed at runtime
COPY --from=builder /app/main ./main
COPY --from=builder /app/config ./config

# Tell the app which config to load (override at run time as needed)
ENV CONFIG_PATH=config/prod/config.yml

# gRPC port
EXPOSE 44044

ENTRYPOINT ["./main"]
