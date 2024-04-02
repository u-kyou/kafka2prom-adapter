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

package httpclient

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/u-kyou/kafka2prom-adapter/pkg/metrics"
	"github.com/u-kyou/kafka2prom-adapter/types"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/prometheus/prometheus/prompb"
)

// Client for the remote write requests.
type RemoteWriteClient struct {
	Client            *http.Client
	RemoteWriteConfig *types.RemoteWriteConfig
}

func NewRemoteWriteClient(client *http.Client, remoteWriteConfig *types.RemoteWriteConfig) *RemoteWriteClient {
	return &RemoteWriteClient{
		Client:            client,
		RemoteWriteConfig: remoteWriteConfig,
	}
}

// RemoteWrite sends a batch of samples to the HTTP endpoint.
func (c *RemoteWriteClient) RemoteWrite(ctx context.Context, req *prompb.WriteRequest) error {
	if req == nil {
		err := fmt.Errorf("couldn't write nil prompb.WriteRequest")
		return err
	}
	data, err := proto.Marshal(req)
	if err != nil {
		return err
	}

	compressed := snappy.Encode(nil, data)

	var resErr error
	// multi endpoints separate by ","
	requestUrls := strings.Split(c.RemoteWriteConfig.Endpoint, ",")
	for _, url := range requestUrls {
		httpReq, err := http.NewRequest("POST", url, bytes.NewReader(compressed))
		if err != nil {
			return err
		}
		httpReq.Header.Add("Content-Encoding", "snappy")
		httpReq.Header.Set("Content-Type", "application/x-protobuf")
		httpReq.Header.Set("User-Agent", "kafka2prom-adatper")
		httpReq = httpReq.WithContext(ctx)

		ctx, cancel := context.WithTimeout(context.Background(), c.RemoteWriteConfig.Timeout)
		defer cancel()

		httpResp, err := c.Client.Do(httpReq.WithContext(ctx))
		metrics.RemoteWrittenTotal.WithLabelValues(url).Inc()
		if err != nil {
			resErr = multierror.Append(resErr, err)
			metrics.RemoteWrittenFailed.WithLabelValues(url).Inc()
			continue
		}

		if httpResp != nil {
			if httpResp.StatusCode/100 != 2 {
				scanner := bufio.NewScanner(io.LimitReader(httpResp.Body, 256))
				line := ""
				if scanner.Scan() {
					line = scanner.Text()
				}
				resErr = multierror.Append(resErr, fmt.Errorf("server returned HTTP status %s: %s", httpResp.Status, line))
				metrics.RemoteWrittenFailed.WithLabelValues(url).Inc()
			}

			httpResp.Body.Close()
		}
	}

	return resErr
}
