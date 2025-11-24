FROM golang:1.25.4-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o server ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/server .
COPY --from=builder /build/web ./web
COPY --from=builder /build/config.yaml.example ./config.yaml.example

ENV CONFIG_PATH=/app/config.yaml
ENV DB_PATH=/app/data/calendar.db

RUN mkdir -p /app/data

EXPOSE 8080

CMD ["./server"]
