FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY ./gateway/go.mod ./gateway/go.sum ./
COPY ./utils /utils
RUN go mod download

COPY ./gateway .
RUN go build -o gateway .

RUN go install github.com/air-verse/air@latest

CMD ["air"]