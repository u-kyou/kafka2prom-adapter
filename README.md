## Abstract
kafka2prom-adapter is a tool which can consume json messages from kafka and remote write metrics data to a prometheus like storage

### Json format
``` Json
{
  "timestamp": "1970-01-01T00:00:00Z",
  "value": "9876543210",
  "name": "up",

  "labels": {
    "__name__": "up",
    "label1": "value1",
    "label2": "value2"
  }
}
```

[prometheus-kafka-adapter](https://github.com/Telefonica/prometheus-kafka-adapter) is a service which receives Prometheus metrics through remote_write, marshal into JSON and sends them into Kafka.
 You can use prometheus-kafka-adapter to produce json messages. 

## Build
### About librdkafka
Prebuilt librdkafka binaries are included with the Go client and librdkafka does not need to be installed separately on the build or target system. The following platforms are supported by the prebuilt librdkafka naries:
- Mac OSX x64 and arm64
- glibc-based Linux x64 and arm64 (e.g., RedHat, Debian, CentOS, Ubuntu, etc) - without GSSAPI/Kerberos support
- musl-based Linux amd64 and arm64 (Alpine) - without GSSAPI/Kerberos support
- Windows amd64 - without GSSAPI/Kerberos support

When building your application for Alpine Linux (musl libc) you must pass -tags musl to go get, go build, etc.
- CGO_ENABLED must NOT be set to 0 since the Go client is based on the C library librdkafka.

If GSSAPI/Kerberos authentication support is required you will need to install librdkafka separately, see the Installing librdkafka chapter below, and then build your Go application with -tags dynamic.
- Installing librdkafka: https://github.com/confluentinc/confluent-kafka-go?tab=readme-ov-file#installing-librdkafka


### go build
```
sh deploy/build.sh cmd/main.go
```

### image build
```
docker  build -t kafka2prom-adapter:v0.0.1 -f deploy/Dockerfile . 
```


## Usage
`kafka2prom-adapter --help` shows help info:

```
usage: kafka2prom-adapter --kafka-broker-list=KAFKA-BROKER-LIST --kafka-topic=KAFKA-TOPIC --kafka-consumer-group-id=KAFKA-CONSUMER-GROUP-ID --remote-write-endpoint=REMOTE-WRITE-ENDPOINT [<flags>]

kafka2prom-adapter - consume json messages from kafka, and remote write metrics data to prometheus like storage

Flags:
  --help                         Show context-sensitive help (also try --help-long and
                                 --help-man).
  --log-level="info"             Specify log level: trace, debug, info, warn, error, fatal,
                                 panic. Default is info.
  --kafka-broker-list=KAFKA-BROKER-LIST
                                 Kafka broker list, separate by comma
  --kafka-topic=KAFKA-TOPIC      Specify kafka topic
  --kafka-ssl-client-cert-file=""
                                 Specify ssl client cert file
  --kafka-ssl-client-key-file=""
                                 Specify ssl client key file
  --kafka-ssl-client-key-pass=""
                                 Specify ssl client key pass
  --kafka-ssl-ca-cert-file=""    Specify ssl ca cert file
  --kafka-security-protocal="ssl"
                                 Specify security protocal: ssl, sasl_ssl or sasl_plaintext.
                                 Default is ssl
  --kafka-sasl-mechanism=""      Specify sasl mechanism
  --kafka-sasl-username=""       Specify sasl username
  --kafka-sasl-password=""       Specify sasl password
  --kafka-consumer-group-id=KAFKA-CONSUMER-GROUP-ID
                                 Specify consumer group id
  --kafka-offset=1               Specify kafka offset. If it is equal to 0, consuming from 
                                 beginning; else from end. If 'multiple-ts' mode is specified, 
                                 the value should be a Unix timestamp in milliseconds. 
                                 Default is to consume from end
  --remote-write-batch-size=100  Specify msg batch size for remote write. Default is 100
  --remote-write-endpoint=REMOTE-WRITE-ENDPOINT
                                 Specify remote write endpoint
  --consumer-instance-mode="multiple"
                                 Specify consumer instance mode: multiple, multiple-ts or single. 
                                 Default is multiple
```

example:
```
kafka2prom-adapter \
--kafka-broker-list \
10.192.10.1:9092,10.192.10.2:9092,10.192.10.3:9092 \
--kafka-topic \
sre_prometheus_remote_write \
--kafka-consumer-group-id \
kafka2prom-consumer \
--remote-write-endpoint \
http://remote-write-address/insert/0/prometheus/ \
--remote-write-batch-size \
200 \
--consumer-instance-mode \
multiple 
```

