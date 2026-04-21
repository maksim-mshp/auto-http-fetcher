FROM golang:1.25.3-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/users ./cmd/users

FROM alpine:3.21
WORKDIR /app
RUN apk add --no-cache ca-certificates git
COPY --from=builder /out/users /app/bin/users
RUN chmod +x /app/bin/users
EXPOSE 8095
CMD ["./bin/users"]