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

	"github.com/booster-proj/booster/core"
	"github.com/gorilla/mux"
)

type StaticInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`

	ProxyPort  int    `json:"proxy_port"`
	ProxyProto string `json:"proxy_proto"`
}

type Router struct {
	r *mux.Router

	Info       StaticInfo
	SourceEnum func(func(core.Source))
}

func NewRouter() *Router {
	return &Router{r: mux.NewRouter()}
}

func (r *Router) SetupRoutes() {
	router := r.r
	router.HandleFunc("/_health", makeHealthCheckHandler(r.Info))
	router.HandleFunc("/sources", makeListSourcesHandler(r.SourceEnum))
	router.Use(loggingMiddleware)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.r.ServeHTTP(w, req)
}
