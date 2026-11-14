FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /vendorsaathi-api ./cmd/api \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /vendorsaathi-migrate ./cmd/migrate \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /vendorsaathi-create-admin ./cmd/create-admin

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=builder --chown=app:app /vendorsaathi-api .
COPY --from=builder --chown=app:app /vendorsaathi-migrate .
COPY --from=builder --chown=app:app /vendorsaathi-create-admin .
COPY --from=builder --chown=app:app /app/migrations ./migrations
COPY --from=builder --chown=app:app /app/config.yaml.example ./config.yaml.example

ENV APP_ENV=production
ENV HTTP_ADDR=:8080

EXPOSE 8080

# Run as an unprivileged user. The API only needs read access to migrations.
USER app

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/health/ready || exit 1

ENTRYPOINT ["./vendorsaathi-api"]