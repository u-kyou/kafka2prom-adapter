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

package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	DeserializeTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "deserialized_total",
			Help: "Count of all deserialization requests",
		})
	DeserializeFailed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "deserialized_failed_total",
			Help: "Count of all deserialization failures",
		})
	RemoteWrittenTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "remote_written_total",
			Help: "Count of all remote written to prometheus",
		},
		[]string{"endpoint"},
	)
	RemoteWrittenFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "remote_written_failed_total",
			Help: "Count of all failed remote written to prometheus",
		},
		[]string{"endpoint"},
	)
	ProcessedMessagesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "processed_message_total",
			Help: "Count of all messages consumed from Kafka",
		},
		[]string{"partition"},
	)
	ConsumerReadMessagesErrorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "consumer_readmessage_error_total",
			Help: "Count of all readmessages error when consuming from Kafka",
		},
		[]string{"partition"},
	)
	ConsumerTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "consumer_goroutine_total",
			Help: "Count of all consumer goroutines",
		})
)

func init() {
	prometheus.MustRegister(DeserializeTotal)
	prometheus.MustRegister(DeserializeFailed)
	prometheus.MustRegister(RemoteWrittenTotal)
	prometheus.MustRegister(RemoteWrittenFailed)
	prometheus.MustRegister(ProcessedMessagesTotal)
	prometheus.MustRegister(ConsumerReadMessagesErrorTotal)
	prometheus.MustRegister(ConsumerTotal)
}
