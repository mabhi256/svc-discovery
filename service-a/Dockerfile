FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY service-a/ ./service-a/
RUN go build -o /service-a ./service-a

FROM alpine:3.23.4
# RUN apk add --no-cache ca-certificates
COPY --from=builder /service-a /service-a
ENTRYPOINT ["/service-a"]