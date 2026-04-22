FROM golang:1.25-alpine
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
WORKDIR /app
COPY . .
CMD ["goose", "-dir", "internal/core/postgres/migrations", "postgres", "postgres://postgres:postgres@postgres:5432/auto-http-fetcher?sslmode=disable", "up"]