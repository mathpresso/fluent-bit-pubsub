FROM 526094755604.dkr.ecr.ap-northeast-2.amazonaws.com/mirror/amazon/aws-for-fluent-bit:2.10.0

COPY files/linux_amd64/pubsub.so /fluent-bit/


CMD ["bash", "-c", "/fluent-bit/bin/fluent-bit -e /fluent-bit/firehose.so -e /fluent-bit/cloudwatch.so -e /fluent-bit/kinesis.so -e /fluent-bit/pubsub.so -c /fluent-bit/etc/fluent-bit.conf"]