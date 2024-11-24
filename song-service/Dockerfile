FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./song-service .
COPY ./utils ../
RUN go build -o song-service ./song-service

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/song-service .

CMD ["./song-service"]