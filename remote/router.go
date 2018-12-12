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

package remote

import (
	"net/http"

	"github.com/booster-proj/booster"
	"github.com/booster-proj/booster/store"
	"github.com/gorilla/mux"
)

type SnapshotProvider interface {
	GetSourcesSnapshot() []*store.DummySource
	GetPoliciesSnapshot() []*store.Policy
}

type Router struct {
	r *mux.Router

	Config   booster.Config
	Provider SnapshotProvider
}

func NewRouter() *Router {
	return &Router{r: mux.NewRouter()}
}

func (r *Router) SetupRoutes() {
	router := r.r
	router.HandleFunc("/_health", makeHealthCheckHandler(r.Config))
	router.HandleFunc("/sources", makeSourcesSnapshotHandler(r.Provider))
	router.HandleFunc("/policies", makePoliciesSnapshotHandler(r.Provider))
	router.Use(loggingMiddleware)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.r.ServeHTTP(w, req)
}
