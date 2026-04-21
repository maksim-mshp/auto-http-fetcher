FROM golang:1.25.3-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/scheduler ./cmd/scheduler

FROM alpine:3.21
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/scheduler /app/bin/scheduler
RUN chmod +x /app/bin/scheduler
CMD ["./bin/scheduler"]
