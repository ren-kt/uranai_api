FROM golang:1.16

WORKDIR /go/src/github.com/ren-kt/uranai_api

RUN apt-get update -qq && apt-get install -y vim git sqlite3 libsqlite3-dev

RUN go install github.com/pilu/fresh@latest
RUN go install golang.org/x/tools/cmd/goimports@latest
RUN go install github.com/kisielk/errcheck@latest

CMD ["fresh"]