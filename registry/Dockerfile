FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY registry/ ./registry/
RUN go build -o /registry ./registry/cmd

FROM alpine:3.23.4
# RUN apk add --no-cache ca-certificates
COPY --from=builder /registry /registry
ENTRYPOINT ["/registry"]