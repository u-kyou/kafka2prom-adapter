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

package prom

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/u-kyou/kafka2prom-adapter/pkg/metrics"
	"github.com/u-kyou/kafka2prom-adapter/types"

	"github.com/prometheus/prometheus/prompb"
	"github.com/sirupsen/logrus"
)

// DeserializeList ...
// jsonDataList example: [{"labels":{"__name__":"prometheus_tsdb_time_retentions_total","instance":"localhost:9090","job":"prometheus"},"name":"prometheus_tsdb_time_retentions_total","timestamp":"2023-10-26T09:19:41Z","value":"0"}]
func DeserializeList(jsonDataList []string) *prompb.WriteRequest {
	var prompbWreq *prompb.WriteRequest
	var timeSerices []prompb.TimeSeries

	for _, jsonData := range jsonDataList {
		metric := &types.JsonMetric{}
		err := json.Unmarshal([]byte(jsonData), metric)
		if err != nil {
			logrus.WithError(err).Errorln("couldn't unmarshal json metric")
			metrics.DeserializeFailed.Inc()
			return nil
		}

		labels := make([]prompb.Label, 0, len(metric.Labels))
		for name, value := range metric.Labels {
			labels = append(labels, prompb.Label{
				Name:  name,
				Value: value,
			})
		}

		value, _ := strconv.ParseFloat(metric.Value, 64)
		t, _ := time.Parse(time.RFC3339, metric.Timestamp)
		timeSerices = append(timeSerices, prompb.TimeSeries{
			Labels: labels,
			Samples: []prompb.Sample{
				{Value: value, Timestamp: t.Unix() * 1000},
			},
		})
	}

	prompbWreq = &prompb.WriteRequest{
		Timeseries: timeSerices,
	}
	metrics.DeserializeTotal.Inc()
	return prompbWreq
}

// Deserialize ...
// jsonData example: {"labels":{"__name__":"prometheus_tsdb_time_retentions_total","instance":"localhost:9090","job":"prometheus"},"name":"prometheus_tsdb_time_retentions_total","timestamp":"2023-10-26T09:19:41Z","value":"0"}
func Deserialize(jsonData string) (*prompb.WriteRequest, error) {
	logrus.WithField("var", jsonData).Debugln()
	var prompbWreq *prompb.WriteRequest

	metric := &types.JsonMetric{}
	err := json.Unmarshal([]byte(jsonData), metric)
	if err != nil {
		logrus.WithError(err).Errorln("couldn't unmarshal json metric")
		return nil, err
	}

	var timeSerices []prompb.TimeSeries

	labels := make([]prompb.Label, 0, len(metric.Labels))
	for name, value := range metric.Labels {
		labels = append(labels, prompb.Label{
			Name:  name,
			Value: value,
		})
	}

	value, _ := strconv.ParseFloat(metric.Value, 64)
	t, _ := time.Parse(time.RFC3339, metric.Timestamp)

	timeSerices = append(timeSerices, prompb.TimeSeries{
		Labels: labels,
		Samples: []prompb.Sample{
			{Value: value, Timestamp: t.Unix() * 1000},
		},
	})

	prompbWreq = &prompb.WriteRequest{
		Timeseries: timeSerices,
	}

	return prompbWreq, nil
}
