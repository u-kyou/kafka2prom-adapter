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

package kafka

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/u-kyou/kafka2prom-adapter/pkg/httpclient"
	"github.com/u-kyou/kafka2prom-adapter/pkg/metrics"
	"github.com/u-kyou/kafka2prom-adapter/pkg/prom"
	"github.com/u-kyou/kafka2prom-adapter/types"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/prometheus/prometheus/prompb"
	"github.com/sirupsen/logrus"
)

func ConsumeKafkaMessagesWithMultipleInstance(ctx context.Context, kafkaConfig *kafka.ConfigMap, kafkaTopic string, kafkaOffset int64, remoteWriteConfig *types.RemoteWriteConfig) {
	// if kafkaOffset is specified
	if kafkaOffset == 0 {
		kafkaConfig.SetKey("auto.offset.reset", "earliest")
	} else {
		kafkaConfig.SetKey("auto.offset.reset", "latest")
	}

	c, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		logrus.WithError(err).Fatal("couldn't create kafka consumer")
	}

	c.Subscribe(kafkaTopic, nil)
	metrics.ConsumerTotal.Inc()

	logrus.WithFields(logrus.Fields{
		"topic": kafkaTopic,
	}).Infoln("start consuming")

	var msgList []string
	var prompbWrep *prompb.WriteRequest

	for {
		select {
		case <-ctx.Done():
			// flush the appended data before exit
			if len(msgList) != 0 {
				prompbWrep := prom.DeserializeList(msgList)
				logrus.Infof("flush %v messages in the appended list", len(msgList))

				if prompbWrep != nil {
					rwc := httpclient.NewRemoteWriteClient(&http.Client{}, remoteWriteConfig)
					remotewriteErr := rwc.RemoteWrite(context.TODO(), prompbWrep)
					if remotewriteErr != nil {
						logrus.WithError(remotewriteErr).Errorln("remotewrite error")
					}
				}
			}
			logrus.Infoln("stopped consuming")
			return
		default:
			msg, readMsgErr := c.ReadMessage(-1)
			if readMsgErr == nil {
				// if the length is less than batch size
				if len(msgList) < remoteWriteConfig.BatchSize {
					msgList = append(msgList, string(msg.Value))

					metrics.ProcessedMessagesTotal.WithLabelValues(fmt.Sprintf("%v", msg.TopicPartition.Partition)).Inc()
					continue
				} else { // Todo: 如果已经poll了一定时间(例如1秒)，则即使没有达到MaxMessages也进行处理. e.g: len(msgList) > 0 && time.Since(msgList[0].Timestamp) > time.Second
					// flush msgList while the length is equal the batch size
					prompbWrep = prom.DeserializeList(msgList)
					msgList = make([]string, 0)
				}
			} else {
				logrus.WithError(readMsgErr).Errorln("consumer ReadMessage error")
				metrics.ConsumerReadMessagesErrorTotal.WithLabelValues(fmt.Sprintf("%v", msg.TopicPartition.Partition)).Inc()

				// flush the appended data
				if len(msgList) != 0 {
					prompbWrep = prom.DeserializeList(msgList)
					msgList = make([]string, 0)
				}
			}

			if prompbWrep != nil {
				rwc := httpclient.NewRemoteWriteClient(&http.Client{}, remoteWriteConfig)
				remotewriteErr := rwc.RemoteWrite(context.TODO(), prompbWrep)
				if remotewriteErr != nil {
					logrus.WithError(remotewriteErr).Errorln("remotewrite error")
				}
			}
		}
	}

}

func ConsumeKafkaMessagesFromTimeStamp(ctx context.Context, kafkaConfig *kafka.ConfigMap, kafkaTopic string, startTimestamp int64, remoteWriteConfig *types.RemoteWriteConfig) {
	c, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		logrus.WithError(err).Fatal("couldn't create kafka consumer")
		return
	}

	err = c.SubscribeTopics([]string{kafkaTopic}, nil)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{"topic": kafkaTopic}).Fatal("Failed to subscribe to topics")
		return
	}

	metadata, err := c.GetMetadata(&kafkaTopic, false, 5000)
	if err != nil {
		logrus.WithError(err).Fatal("GetMetadata failed")
		return
	}

	var partitionsWithTimestamps []kafka.TopicPartition
	for _, p := range metadata.Topics[kafkaTopic].Partitions {
		tp := kafka.TopicPartition{Topic: &kafkaTopic, Partition: p.ID}
		partitionsWithTimestamps = append(partitionsWithTimestamps, kafka.TopicPartition{Topic: tp.Topic, Partition: tp.Partition, Offset: kafka.Offset(startTimestamp)})
	}

	// looks up offsets by timestamp for the given partitions
	partitionWithOffsets, err := c.OffsetsForTimes(partitionsWithTimestamps, 5000)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get offsets by OffsetsForTimes")
		return
	}
	c.Assign(partitionWithOffsets)

	defer c.Close()

	var msgList []string
	var prompbWrep *prompb.WriteRequest

	for {
		select {
		case <-ctx.Done():
			// flush the appended data before exit
			if len(msgList) != 0 {
				prompbWrep := prom.DeserializeList(msgList)
				logrus.Infof("flush %v messages in the appended list", len(msgList))

				if prompbWrep != nil {
					rwc := httpclient.NewRemoteWriteClient(&http.Client{}, remoteWriteConfig)
					remotewriteErr := rwc.RemoteWrite(context.TODO(), prompbWrep)
					if remotewriteErr != nil {
						logrus.WithError(remotewriteErr).Errorln("remotewrite error")
					}
				}
			}
			logrus.Infoln("stopped consuming")
			return
		default:
			msg, readMsgErr := c.ReadMessage(-1)
			if readMsgErr == nil {
				// if the length is less than batch size
				if len(msgList) < remoteWriteConfig.BatchSize {
					msgList = append(msgList, string(msg.Value))

					metrics.ProcessedMessagesTotal.WithLabelValues(fmt.Sprintf("%v", msg.TopicPartition.Partition)).Inc()
					continue
				} else { // Todo: 如果已经poll了一定时间(例如1秒)，则即使没有达到MaxMessages也进行处理. e.g: len(msgList) > 0 && time.Since(msgList[0].Timestamp) > time.Second
					// flush msgList while the length is equal the batch size
					prompbWrep = prom.DeserializeList(msgList)
					msgList = make([]string, 0)
				}
			} else {
				logrus.WithError(readMsgErr).Errorln("consumer ReadMessage error")
				metrics.ConsumerReadMessagesErrorTotal.WithLabelValues(fmt.Sprintf("%v", msg.TopicPartition.Partition)).Inc()

				// flush the appended data
				if len(msgList) != 0 {
					prompbWrep = prom.DeserializeList(msgList)
					msgList = make([]string, 0)
				}
			}

			if prompbWrep != nil {
				rwc := httpclient.NewRemoteWriteClient(&http.Client{}, remoteWriteConfig)
				remotewriteErr := rwc.RemoteWrite(context.TODO(), prompbWrep)
				if remotewriteErr != nil {
					logrus.WithError(remotewriteErr).Errorln("remotewrite error")
				}
			}
		}
	}

}

func ConsumeKafkaMessagesWithSingleInstance(ctx context.Context, kafkaConfig *kafka.ConfigMap, kafkaTopic string, kafkaOffset int64, remoteWriteConfig *types.RemoteWriteConfig) {
	var wg sync.WaitGroup

	consumer, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		logrus.WithError(err).Fatal("couldn't create kafka consumer")
	}

	metadata, err := consumer.GetMetadata(&kafkaTopic, false, 5000)
	if err != nil {
		panic(err)
	}
	consumer.Close()

	for _, v := range metadata.Topics[kafkaTopic].Partitions {
		logrus.WithFields(logrus.Fields{
			"topic":     kafkaTopic,
			"partition": v.ID,
		}).Infoln("start consuming")
		wg.Add(1)
		// if kafkaOffset is specified
		var currentOffset kafka.Offset
		if kafkaOffset == 0 {
			currentOffset = kafka.OffsetBeginning
		} else {
			currentOffset = kafka.OffsetEnd
		}

		// create consumer for each partition
		c, err := kafka.NewConsumer(kafkaConfig)
		if err != nil {
			logrus.WithError(err).Fatal("couldn't create kafka consumer")
		}
		defer c.Close()
		// assign partition to consumer
		partitionID := v.ID
		assignErr := c.Assign([]kafka.TopicPartition{{Topic: &kafkaTopic, Partition: partitionID, Offset: currentOffset}})
		if assignErr != nil {
			logrus.WithError(assignErr).Fatal("kafka consumer assign offset error")
		}
		metrics.ConsumerTotal.Inc()

		// process messages in assigned partition
		go ProcessMessages(ctx, c, remoteWriteConfig, partitionID, &wg)
	}
	wg.Wait()
	logrus.Infoln("stopped all kafka messages consumer")
}

func ProcessMessages(ctx context.Context, c *kafka.Consumer, remoteWriteConfig *types.RemoteWriteConfig, partitionID int32, wg *sync.WaitGroup) {
	defer wg.Done()
	var msgList []string
	for {
		select {
		case <-ctx.Done():
			// flush the appended data before exit
			if len(msgList) != 0 {
				prompbWrep := prom.DeserializeList(msgList)
				logrus.Infof("flush %v messages of partition %v in the appended list", len(msgList), partitionID)

				if prompbWrep != nil {
					rwc := httpclient.NewRemoteWriteClient(&http.Client{}, remoteWriteConfig)
					remotewriteErr := rwc.RemoteWrite(context.TODO(), prompbWrep)
					if remotewriteErr != nil {
						logrus.WithError(remotewriteErr).Errorln("remotewrite error")
					}
				}
			}
			logrus.Infof("stopped consuming partition: %v", partitionID)
			return
		default:
			var prompbWrep *prompb.WriteRequest

			msg, readMsgErr := c.ReadMessage(-1)
			if readMsgErr == nil {
				// if the length is less than batch size
				if len(msgList) < remoteWriteConfig.BatchSize {
					msgList = append(msgList, string(msg.Value))
					metrics.ProcessedMessagesTotal.WithLabelValues(fmt.Sprintf("%v", partitionID)).Inc()
					continue
				} else {
					// flush msgList while the length is equal to the batch size
					prompbWrep = prom.DeserializeList(msgList)
					msgList = make([]string, 0)
				}
			} else {
				logrus.WithError(readMsgErr).Errorln("consumer ReadMessage error")
				metrics.ConsumerReadMessagesErrorTotal.WithLabelValues(fmt.Sprintf("%v", partitionID)).Inc()

				// flush the appended data
				if len(msgList) != 0 {
					prompbWrep = prom.DeserializeList(msgList)
					metrics.ProcessedMessagesTotal.WithLabelValues(fmt.Sprintf("%v", partitionID)).Add(float64(len(msgList)))
					msgList = make([]string, 0)
				}
			}

			if prompbWrep != nil {
				rwc := httpclient.NewRemoteWriteClient(&http.Client{}, remoteWriteConfig)
				remotewriteErr := rwc.RemoteWrite(context.TODO(), prompbWrep)
				if remotewriteErr != nil {
					logrus.WithError(remotewriteErr).Errorln("remotewrite error")
				}
			}
		}
	}
}
