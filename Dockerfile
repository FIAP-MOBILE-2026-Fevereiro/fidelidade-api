FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /bin/api /usr/local/bin/api
COPY db/migrations ./db/migrations
COPY openapi.yaml ./openapi.yaml

EXPOSE 3000

CMD ["api"]