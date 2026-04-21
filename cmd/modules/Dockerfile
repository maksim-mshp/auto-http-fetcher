FROM golang:1.25.3-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/modules ./cmd/modules

FROM alpine:3.21
WORKDIR /app
RUN apk add --no-cache ca-certificates git
COPY --from=builder /out/modules /app/bin/modules
RUN chmod +x /app/bin/modules
EXPOSE 8090
CMD ["./bin/modules"]