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

package metrics

import (
	"github.com/booster-proj/booster/source"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type Broker struct {
}

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

func (b *Broker) SendDataFlow(labels map[string]string, data *source.DataFlow) {
	switch data.Type {
	case "read":
		receiveBytes.With(prometheus.Labels(labels)).Add(float64(data.N))
	case "write":
		sendBytes.With(prometheus.Labels(labels)).Add(float64(data.N))
	default:
	}
}
