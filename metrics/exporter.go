// Copyright © 2019 KIM KeepInMind GmbH/srl
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

// Package metrics provides what in prometheus terms is called a
// metrics exporter.
package metrics

import (
	"net/http"
	"time"

	"github.com/booster-proj/booster/source"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "booster"

var (
	sendBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "network_send_bytes",
		Help:      "Sent bytes for network source",
	}, []string{"source", "target"})

	receiveBytes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "network_receive_bytes",
		Help:      "Received bytes for network source",
	}, []string{"source", "target"})

	selectSource = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "select_source_total",
		Help:      "Number of times a source was chosen",
	}, []string{"source", "target"})

	countConn = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "open_conn_count",
		Help:      "Number of open connections",
	}, []string{"source", "target"})

	addLatency = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "conn_latency_ms",
		Help:      "Latency value measured in milliseconds",
	}, []string{"source", "target"})

	countPort = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "port_count",
		Help:      "Number of times a port is being used",
	}, []string{"port", "protocol"})
)

func init() {
	prometheus.MustRegister(sendBytes)
	prometheus.MustRegister(receiveBytes)
	prometheus.MustRegister(selectSource)
	prometheus.MustRegister(countConn)
	prometheus.MustRegister(addLatency)
	prometheus.MustRegister(countPort)
}

// Exporter can be used to both capture and serve metrics.
type Exporter struct {
}

// ServeHTTP is just a wrapper around the ServeHTTP function
// of the prohttp default Handler.
func (exp *Exporter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}

// SendDataFlow can be used to update the metrics exported by the broker
// about network usage, in particular upload and download bandwidth. `data`
// Type should either be "read" or "write", referring respectively to download
// and upload operations.
func (exp *Exporter) SendDataFlow(labels map[string]string, data *source.DataFlow) {
	switch data.Type {
	case "read":
		receiveBytes.With(prometheus.Labels(labels)).Add(float64(data.N))
	case "write":
		sendBytes.With(prometheus.Labels(labels)).Add(float64(data.N))
	default:
	}
}

// IncSelectedSource is used to update the number of times a source was
// chosen.
func (exp *Exporter) IncSelectedSource(labels map[string]string) {
	selectSource.With(prometheus.Labels(labels)).Inc()
}

// CountOpenConn is used to updated the number of open connections created
// through booster sources.
func (exp *Exporter) CountOpenConn(labels map[string]string, val int) {
	countConn.With(prometheus.Labels(labels)).Add(float64(val))
}

// AddLatency is used to update the latency of the connections opened.
func (exp *Exporter) AddLatency(labels map[string]string, d time.Duration) {
	ms := float64(d / 1000000)
	addLatency.With(prometheus.Labels(labels)).Add(ms)
}

//IncProtocol increments the protocol counter
func (exp *Exporter) IncPortCount(labels map[string]string) {
	countPort.With(prometheus.Labels(labels)).Inc()
}
