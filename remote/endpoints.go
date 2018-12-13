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
	"io"
	"net/http"

	"github.com/booster-proj/booster"
	"github.com/booster-proj/booster/store"
	"github.com/gorilla/mux"
	"upspin.io/log"
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

func makeSourcesHandler(s *store.SourceStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(struct {
			Sources []*store.DummySource `json:"sources"`
		}{
			Sources: s.GetSourcesSnapshot(),
		})
	}
}

func makePoliciesHandler(s *store.SourceStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(struct {
			Policies []*store.Policy `json:"policies"`
		}{
			Policies: s.GetPoliciesSnapshot(),
		})
	}
}

func makeBlockHandler(s *store.SourceStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]

		p := &store.Policy{
			ID:     "block_" + name,
			Issuer: "remote",
			Code:   store.PolicyBlock,
			Accept: func(n string) bool {
				return n != name
			},
		}

		if r.Method == "POST" {
			// Add a reason if available in the body.
			var payload struct {
				Reason string `json:"reason"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil && err != io.EOF {
				log.Error.Printf("block handler: unable to decode body: %v", err)
			}
			r.Body.Close()
			if payload.Reason != "" {
				p.Reason = payload.Reason
			}
			s.AddPolicy(p)
		} else {
			// Only POST and DELETE are registered.
			s.DelPolicy(p.ID)
		}

		w.WriteHeader(http.StatusOK)
	}
}
