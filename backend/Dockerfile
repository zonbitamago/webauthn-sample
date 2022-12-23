FROM golang:1.19.4-alpine3.17 as builder

WORKDIR /go/src
COPY . .

# Airをインストールし、コンテナ起動時に実行する
RUN go install github.com/cosmtrek/air@latest
CMD ["air"]