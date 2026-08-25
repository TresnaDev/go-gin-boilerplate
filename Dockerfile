# --- build stage ---
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum* ./

COPY . .

RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api

# --- run stage ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/bin/api .

EXPOSE 8080

CMD ["./api"]
