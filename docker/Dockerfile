FROM golang:1.22 as gobuilder

RUN apt-get update && apt-get install -y ca-certificates libgnutls30

WORKDIR /build
COPY . .

RUN go build -buildmode=c-shared -o pubsub.so .

FROM cr.fluentbit.io/fluent/fluent-bit:3.2.10

COPY --from=gobuilder /build/pubsub.so /fluent-bit/

COPY conf/ /fluent-bit/parsers

CMD ["/fluent-bit/bin/fluent-bit", "-e", "/fluent-bit/pubsub.so", "-c", "/fluent-bit/etc/fluent-bit.conf"]

