# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.25-alpine AS builder

WORKDIR /src

# Cache deps first.
COPY api/go.mod api/go.sum ./
RUN go mod download

# Build. CGO is disabled so the binary is fully static and portable.
COPY api/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/planetary ./cmd/planetary/

# ---- runtime stage ----
FROM alpine:3.20

# ca-certificates is needed for outbound HTTPS (feed fetching, Limen).
# tzdata gives us zone info for timestamps.
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /out/planetary /app/planetary

# Non-root user for the HTTP server.
RUN adduser -D -u 10001 planetary \
 && chown -R planetary:planetary /app
USER planetary

EXPOSE 8080

ENTRYPOINT ["/app/planetary"]
