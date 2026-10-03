# --- Build Stage ---
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /singbox-rule-cache .

# --- Runtime Stage ---
FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder /singbox-rule-cache /usr/local/bin/singbox-rule-cache
USER appuser
HEALTHCHECK --interval=30s --timeout=10s --retries=3 \
  CMD singbox-rule-cache list || exit 1
ENTRYPOINT ["singbox-rule-cache"]
CMD ["start"]
