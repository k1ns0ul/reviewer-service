FROM golang:1.25.3-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/reviewer-service ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates netcat-openbsd

WORKDIR /app

COPY --from=builder /app/bin/reviewer-service .
COPY scripts/wait-for-it.sh ./wait-for-it.sh
RUN chmod +x wait-for-it.sh

EXPOSE 8080

CMD ["./reviewer-service"]
