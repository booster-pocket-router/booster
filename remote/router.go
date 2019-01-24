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

package remote

import (
	"net/http"

	"github.com/booster-proj/booster/store"
	"github.com/gorilla/mux"
)

type BoosterInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`

	ProxyPort  int    `json:"proxy_port"`
	ProxyProto string `json:"proxy_proto"`
	PromPort   int    `json:"-"`
}

var Info BoosterInfo = BoosterInfo{}

type Router struct {
	r *mux.Router

	Store           *store.SourceStore
	MetricsProvider http.Handler
}

func NewRouter() *Router {
	return &Router{r: mux.NewRouter()}
}

// SetupRoutes adds the routes available to the router. Make sure
// to fill the public fields of the Router before calling this
// function, otherwise the handlers will not be able to work
// properly.
func (r *Router) SetupRoutes() {
	router := r.r
	router.HandleFunc("/health.json", healthCheckHandler)
	if store := r.Store; store != nil {
		router.HandleFunc("/sources.json", makeSourcesHandler(store))
		router.HandleFunc("/sources/{name}/block.json", makeBlockHandler(store)).Methods("POST", "DELETE")
		router.HandleFunc("/policies.json", makePoliciesHandler(store))
	}
	if handler := r.MetricsProvider; handler != nil {
		router.Handle("/metrics", handler)
	}
	router.HandleFunc("/metrics.json", metricsForwardHandler)
	router.Use(loggingMiddleware)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.r.ServeHTTP(w, req)
}
