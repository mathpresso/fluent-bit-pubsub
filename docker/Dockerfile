FROM golang:1.12 as gobuilder

ENV GO111MODULE on

RUN apt update && apt install ca-certificates libgnutls30 -y

COPY . /go/pubsub/

WORKDIR /go/pubsub

RUN go build -buildmode=c-shared -o pubsub.so .

FROM 526094755604.dkr.ecr.ap-northeast-2.amazonaws.com/mirror/amazon/aws-for-fluent-bit:2.10.0

COPY --from=gobuilder /go/pubsub/pubsub.so /fluent-bit/


CMD ["bash", "-c", "/fluent-bit/bin/fluent-bit -e /fluent-bit/firehose.so -e /fluent-bit/cloudwatch.so -e /fluent-bit/kinesis.so -e /fluent-bit/pubsub.so -c /fluent-bit/etc/fluent-bit.conf"]
