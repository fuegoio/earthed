# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Cache deps first.
COPY go/relay/go.mod go/relay/go.sum ./
RUN go mod download

# Build. CGO is disabled so the binary is fully static and portable.
COPY go/relay/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/sunred-relay ./cmd/relay/

# ---- runtime stage ----
FROM alpine:3.20

# ca-certificates is needed for outbound HTTPS (PDS subscriptions).
# tzdata gives us zone info for timestamps.
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /out/sunred-relay /app/sunred-relay

# Non-root user for the HTTP server.
RUN adduser -D -u 10001 sunred \
 && chown -R sunred:sunred /app
USER sunred

EXPOSE 9090

ENTRYPOINT ["/app/sunred-relay"]
