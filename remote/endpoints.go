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
	"encoding/json"
	"net/http"

	"github.com/booster-proj/booster"
	"github.com/booster-proj/booster/store"
)

func makeHealthCheckHandler(config booster.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(struct {
			Alive bool `json:"alive"`
			booster.Config
		}{
			Alive:  true,
			Config: config,
		})
	}
}

func makeSourcesSnapshotHandler(p SnapshotProvider) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(struct {
			Sources []*store.DummySource `json:"sources"`
		}{
			Sources: p.GetSourcesSnapshot(),
		})
	}
}

func makePoliciesSnapshotHandler(p SnapshotProvider) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(struct {
			Policies []*store.Policy `json:"policies"`
		}{
			Policies: p.GetPoliciesSnapshot(),
		})
	}
}
