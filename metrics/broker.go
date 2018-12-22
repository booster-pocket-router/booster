/*
Copyright (C) 2018 KIM KeepInMind GmbH/srl

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

// Package metrics provides what in prometheus terms is called a
// metrics exporter.
package metrics

import (
	"net/http"

	"github.com/booster-proj/booster/source"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Broker can be used to both capture and serve metrics.
type Broker struct {
}

// ServeHTTP is just a wrapper around the ServeHTTP function
// of the prohttp default Handler.
func (b *Broker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}

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
)

func init() {
	prometheus.MustRegister(sendBytes)
	prometheus.MustRegister(receiveBytes)
}

// SendDataFlow can be used to update the metrics exported by the broker
// about network usage, in particular upload and download bandwidth. `data`
// Type should either be "read" or "write", referring respectively to download
// and upload operations.
func (b *Broker) SendDataFlow(labels map[string]string, data *source.DataFlow) {
	switch data.Type {
	case "read":
		receiveBytes.With(prometheus.Labels(labels)).Add(float64(data.N))
	case "write":
		sendBytes.With(prometheus.Labels(labels)).Add(float64(data.N))
	default:
	}
}
