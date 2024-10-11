FROM golang:1.23.1-alpine3.20 as builder

WORKDIR /go/src
COPY . .

# Airをインストールし、コンテナ起動時に実行する
RUN go install github.com/air-verse/air@latest
CMD ["air"]