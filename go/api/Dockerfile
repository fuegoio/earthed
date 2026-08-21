# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.26-alpine AS builder

WORKDIR /src

# Cache deps first.
COPY go/api/go.mod go/api/go.sum ./
RUN go mod download

# Build. CGO is disabled so the binary is fully static and portable.
COPY go/api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/sunred ./cmd/sunred/

# ---- runtime stage ----
FROM alpine:3.20

# ca-certificates is needed for outbound HTTPS (feed fetching, Limen).
# tzdata gives us zone info for timestamps.
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /out/sunred /app/sunred

# Non-root user for the HTTP server.
RUN adduser -D -u 10001 sunred \
 && chown -R sunred:sunred /app
USER sunred

EXPOSE 8080

ENTRYPOINT ["/app/sunred"]
