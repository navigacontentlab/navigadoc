FROM golang:1.13

ADD vendor /go/src

RUN go get github.com/dvyukov/go-fuzz/go-fuzz \
    github.com/dvyukov/go-fuzz/go-fuzz-build

ENV GO111MODULE off

RUN useradd -ms /bin/bash go-fuzz
