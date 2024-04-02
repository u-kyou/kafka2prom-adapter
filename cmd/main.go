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

package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/u-kyou/kafka2prom-adapter/pkg/config"
	kfk "github.com/u-kyou/kafka2prom-adapter/pkg/kafka"

	"github.com/sirupsen/logrus"
)

func main() {
	// start consuming kafka messages
	ctx, cancel := context.WithCancel(context.Background())
	consumerInstanceMode := config.GetConsumerInstanceMode()
	switch consumerInstanceMode {
	case "multiple": // multiple consumer instances(each instance consume a partition)
		go kfk.ConsumeKafkaMessagesWithMultipleInstance(
			ctx, config.NewKafkaConfig(),
			config.GetKafkaTopic(),
			config.GetKafkaOffset(),
			config.NewRemoteWriteConfig())
	case "multiple-ts": // Support consumption from a specified timestamp. Single consumer instance, but multiple consumer goroutines(each goroutine consume a partition)
		go kfk.ConsumeKafkaMessagesFromTimeStamp(
			ctx, config.NewKafkaConfig(),
			config.GetKafkaTopic(),
			config.GetKafkaOffset(),
			config.NewRemoteWriteConfig())
	default: // single consumer instance, but multiple consumer goroutines(each goroutine consume a partition)
		go kfk.ConsumeKafkaMessagesWithSingleInstance(
			ctx, config.NewKafkaConfig(),
			config.GetKafkaTopic(),
			config.GetKafkaOffset(),
			config.NewRemoteWriteConfig())
	}

	// config http server for metrics
	srv := http.Server{Addr: ":8081"}
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// signal channel for stop
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		logrus.WithFields(logrus.Fields{
			"signal": sig,
		}).Infoln("stop consume kafka topic and exit")
		// stop consuming all kafka partitions
		cancel()
		// wait consumers
		time.Sleep(3 * time.Second)

		// stop http server
		if err := srv.Shutdown(context.TODO()); err != nil {
			logrus.WithError(err).Fatalln("http server shutdown error")
		}
		logrus.Infoln("stopped http server")
		logrus.Infoln("see you next time!")
	}()

	// run http server
	logrus.Infoln("run http server for metrics and healthz")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		logrus.Fatalf("ListenAndServe(): %v", err)
	}
}
