FROM golang:1.24.1-alpine3.21

RUN apk add --no-cache inotify-tools gcc g++ musl-dev tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . /app

RUN go install github.com/air-verse/air@latest

ARG TZ=UTC
ENV TZ=$TZ

CMD air