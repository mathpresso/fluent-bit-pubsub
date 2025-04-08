FROM golang:1.12 as gobuilder

ENV GO111MODULE on

RUN apt update && apt install ca-certificates libgnutls30 -y

COPY . /go/pubsub/

WORKDIR /go/pubsub

RUN go build -buildmode=c-shared -o pubsub.so .

FROM cr.fluentbit.io/fluent/fluent-bit:3.2.10

COPY --from=gobuilder /go/pubsub/pubsub.so /fluent-bit/

COPY conf/ /fluent-bit/parsers

CMD ["/fluent-bit/bin/fluent-bit", "-e", "/fluent-bit/pubsub.so", "-c", "/fluent-bit/etc/fluent-bit.conf"]

