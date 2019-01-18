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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"compress/gzip"

	"github.com/booster-proj/booster/store"
	"github.com/gorilla/mux"
	"upspin.io/log"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(struct {
		Alive bool `json:"alive"`
		Config
	}{
		Alive:  true,
		Config: StaticConf,
	})
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

		w.Header().Set("Content-Type", "application/json")

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
			w.WriteHeader(http.StatusCreated)
		} else {
			// Only POST and DELETE are registered.
			s.DelPolicy(p.ID)
			w.WriteHeader(http.StatusOK)
		}
	}
}

func metricsForwardHandler(w http.ResponseWriter, r *http.Request) {
	URL, _ := url.Parse(r.URL.String())
	URL.Scheme = "http"
	URL.Host = fmt.Sprintf("localhost:%d", StaticConf.PromPort)
	URL.Path = "api/v1/query"

	req, err := http.NewRequest(r.Method, URL.String(), r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	req.Header = r.Header

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	gzipR, err := gzip.NewReader(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, gzipR)
}
