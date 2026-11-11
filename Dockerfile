FROM golang:1.27.4-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static" -X main.Version='${VERSION} -o server .

FROM alpine:latest
LABEL org.opencontainers.image.title="linke-calendar" \
  org.opencontainers.image.description="Go service scraping Zetkin events and serving an embeddable calendar" \
  org.opencontainers.image.vendor="Roman Zipp" \
  org.opencontainers.image.authors="Roman Zipp" \
  org.opencontainers.image.licenses="MIT" \
  org.opencontainers.image.url="https://codeberg.org/romanzipp/linke-calendar" \
  org.opencontainers.image.source="https://codeberg.org/romanzipp/linke-calendar" \
  org.opencontainers.image.documentation="https://codeberg.org/romanzipp/linke-calendar"

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
