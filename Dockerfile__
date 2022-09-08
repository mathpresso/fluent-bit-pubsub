FROM golang:1.12

ENV GO111MODULE on

RUN apt update && apt install ca-certificates libgnutls30 -y

COPY . /go/pubsub/

WORKDIR /go/pubsub

RUN go build -buildmode=c-shared -o pubsub.so .
