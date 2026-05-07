FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY service-b/ ./service-b/
RUN go build -o /service-b ./service-b

FROM alpine:3.23.4
# RUN apk add --no-cache ca-certificates
COPY --from=builder /service-b /service-b
ENTRYPOINT ["/service-b"]