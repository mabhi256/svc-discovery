FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY service-x/ ./service-x/
RUN go build -o /service-x ./service-x/cmd

FROM alpine:3.23.4
# RUN apk add --no-cache ca-certificates
COPY --from=builder /service-x /service-x
ENTRYPOINT ["/service-x"]