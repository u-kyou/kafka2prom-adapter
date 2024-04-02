// Copyright 2024 u-kyou
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"os"
	"time"

	"github.com/u-kyou/kafka2prom-adapter/types"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	logLevel                   = kingpin.Flag("log-level", "Specify log level: trace, debug, info, warn, error, fatal, panic. Default is info.").Default("info").String()
	kafkaBrokerList            = kingpin.Flag("kafka-broker-list", "Kafka broker list, separate by comma").Required().String()
	kafkaTopic                 = kingpin.Flag("kafka-topic", "Specify kafka topic").Required().String()
	kafkaSslClientCertFile     = kingpin.Flag("kafka-ssl-client-cert-file", "Specify ssl client cert file").Default("").String()
	kafkaSslClientKeyFile      = kingpin.Flag("kafka-ssl-client-key-file", "Specify ssl client key file").Default("").String()
	kafkaSslClientKeyPass      = kingpin.Flag("kafka-ssl-client-key-pass", "Specify ssl client key pass").Default("").String()
	kafkaSslCACertFile         = kingpin.Flag("kafka-ssl-ca-cert-file", "Specify ssl ca cert file").Default("").String()
	kafkaSecurityProtocol      = kingpin.Flag("kafka-security-protocal", "Specify security protocal: ssl, sasl_ssl or sasl_plaintext. Default is ssl").Default("ssl").String()
	kafkaSaslMechanism         = kingpin.Flag("kafka-sasl-mechanism", "Specify sasl mechanism").Default("").String()
	kafkaSaslUsername          = kingpin.Flag("kafka-sasl-username", "Specify sasl username").Default("").String()
	kafkaSaslPassword          = kingpin.Flag("kafka-sasl-password", "Specify sasl password").Default("").String()
	kafkaConsumerGroupId       = kingpin.Flag("kafka-consumer-group-id", "Specify consumer group id").Required().String()
	kafkaOffset                = kingpin.Flag("kafka-offset", "Specify kafka offset. If it is equal to 0, consuming from beginning; else from end. If 'multiple-ts' mode is specified, the value should be a Unix timestamp in milliseconds. Default is to consume from end").Default("1").Int64()
	msgBatchSizeForRemoteWrite = kingpin.Flag("remote-write-batch-size", "Specify msg batch size for remote write. Default is 100").Default("100").Int()
	remoteWriteEndpoint        = kingpin.Flag("remote-write-endpoint", "Specify remote write endpoint").Required().String()
	consumerInstanceMode       = kingpin.Flag("consumer-instance-mode", "Specify consumer instance mode: multiple, multiple-ts or single. Default is multiple").Default("multiple").String()

	remoteWriteTimeout = time.Second * 5

	Version   = "Unkown"
	Commit    = "Unkown"
	BuildDate = "Unkown"
)

func init() {
	kingpin.CommandLine.Help = "kafka2prom-adapter - consume json messages from kafka, and remote write metrics data to prometheus like storage"
	versionInfo := fmt.Sprintf("version: %v\ncommit: %v\nbuild date: %v\n", Version, Commit, BuildDate)
	kingpin.CommandLine.Version(versionInfo)
	kingpin.Parse()

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	if *logLevel != "" {
		logrus.SetLevel(parseLogLevel(*logLevel))
	}
}

func parseLogLevel(value string) logrus.Level {
	level, err := logrus.ParseLevel(value)

	if err != nil {
		logrus.WithField("log-level-value", value).Warningln("invalid log level from env var, using info")
		return logrus.InfoLevel
	}
	return level
}

func NewKafkaConfig() *kafka.ConfigMap {
	logrus.Info("creating kafka ConfigMap...")

	kafkaConfig := kafka.ConfigMap{
		"bootstrap.servers": *kafkaBrokerList,
		"group.id":          *kafkaConsumerGroupId,
		"auto.offset.reset": "latest",
		"fetch.min.bytes":   50000,
		// "fetch.max.bytes":   1048576, // 1 MB
	}

	if *kafkaSslClientCertFile != "" && *kafkaSslClientKeyFile != "" && *kafkaSslCACertFile != "" {
		if *kafkaSecurityProtocol != "ssl" && *kafkaSecurityProtocol != "sasl_ssl" {
			logrus.Fatal("invalid config: kafka security protocol is not ssl based but ssl config is provided")
		}

		kafkaConfig["security.protocol"] = kafkaSecurityProtocol
		kafkaConfig["ssl.ca.location"] = kafkaSslCACertFile              // CA certificate file for verifying the broker's certificate.
		kafkaConfig["ssl.certificate.location"] = kafkaSslClientCertFile // Client's certificate
		kafkaConfig["ssl.key.location"] = kafkaSslClientKeyFile          // Client's key
		kafkaConfig["ssl.key.password"] = kafkaSslClientKeyPass          // Key password, if any.
	}

	if *kafkaSaslMechanism != "" && *kafkaSaslUsername != "" && *kafkaSaslPassword != "" {
		if *kafkaSecurityProtocol != "sasl_ssl" && *kafkaSecurityProtocol != "sasl_plaintext" {
			logrus.Fatal("invalid config: kafka security protocol is not sasl based but sasl config is provided")
		}

		kafkaConfig["security.protocol"] = kafkaSecurityProtocol
		kafkaConfig["sasl.mechanism"] = kafkaSaslMechanism
		kafkaConfig["sasl.username"] = kafkaSaslUsername
		kafkaConfig["sasl.password"] = kafkaSaslPassword
	}
	return &kafkaConfig
}

func NewRemoteWriteConfig() *types.RemoteWriteConfig {
	return &types.RemoteWriteConfig{
		Endpoint:  *remoteWriteEndpoint,
		BatchSize: *msgBatchSizeForRemoteWrite,
		Timeout:   remoteWriteTimeout,
	}
}

func GetKafkaTopic() string {
	return *kafkaTopic
}

func GetKafkaOffset() int64 {
	return *kafkaOffset
}

func GetConsumerInstanceMode() string {
	return *consumerInstanceMode
}
