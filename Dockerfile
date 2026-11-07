FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /vendorsaathi-api ./cmd/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /vendorsaathi-api .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/config.yaml.example ./config.yaml.example

ENV APP_ENV=production
ENV HTTP_ADDR=:8080

EXPOSE 8080

ENTRYPOINT ["./vendorsaathi-api"]