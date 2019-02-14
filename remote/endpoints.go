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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/booster-proj/booster/store"
	"github.com/gorilla/mux"
)

func makeHealthCheckHandler(info BoosterInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(struct {
			Alive bool `json:"alive"`
			BoosterInfo
		}{
			Alive:       true,
			BoosterInfo: info,
		})
	}
}

func makeSourcesHandler(s *store.SourceStore) http.HandlerFunc {
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

func makePoliciesHandler(s *store.SourceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(struct {
			Policies []store.Policy `json:"policies"`
		}{
			Policies: s.GetPoliciesSnapshot(),
		})
	}
}

func makePoliciesDelHandler(s *store.SourceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		err := s.DelPolicy(id)
		if err != nil {
			writeError(w, err, http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// PoliciesInput describes the fields required by most `POST` requests
// to a `/policies/...` endpoint.
type PoliciesInput struct {
	SourceID string `json:"source_id"`
	Target   string `json:"target"`
	Reason   string `json:"reason"`
	Issuer   string `json:"issuer"`
}

func makePoliciesBlockHandler(s *store.SourceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload PoliciesInput
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, err, http.StatusBadRequest)
			return
		}
		if payload.SourceID == "" {
			writeError(w, fmt.Errorf("validation error: source_id cannot be empty"), http.StatusBadRequest)
			return
		}

		p := store.NewBlockPolicy(payload.Issuer, payload.SourceID)
		p.Reason = payload.Reason
		handlePolicy(s, p, w, r)
	}
}

func makePoliciesStickyHandler(s *store.SourceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload PoliciesInput
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, err, http.StatusBadRequest)
			return
		}

		p := store.NewStickyPolicy(payload.Issuer, s.QueryBindHistory)
		handlePolicy(s, p, w, r)
	}
}

type ReservedPolicyInput struct {
	PoliciesInput
	Hosts []string `json:"hosts"`
}

func makePoliciesReserveHandler(s *store.SourceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload ReservedPolicyInput
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, err, http.StatusBadRequest)
			return
		}
		if payload.SourceID == "" {
			writeError(w, fmt.Errorf("validation error: source_id cannot be empty"), http.StatusBadRequest)
			return
		}
		if len(payload.Hosts) == 0 {
			writeError(w, fmt.Errorf("validation error: hosts cannot be empty list"), http.StatusBadRequest)
			return
		}

		p := store.NewReservedPolicy(payload.Issuer, payload.SourceID, payload.Hosts...)
		p.Reason = payload.Reason
		handlePolicy(s, p, w, r)
	}
}

func makePoliciesAvoidHandler(s *store.SourceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload PoliciesInput
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, err, http.StatusBadRequest)
			return
		}
		if payload.SourceID == "" {
			writeError(w, fmt.Errorf("validation error: source_id cannot be empty"), http.StatusBadRequest)
			return
		}
		if payload.Target == "" {
			writeError(w, fmt.Errorf("validation error: target cannot be empty"), http.StatusBadRequest)
			return
		}

		p := store.NewAvoidPolicy(payload.Issuer, payload.SourceID, payload.Target)
		p.Reason = payload.Reason
		handlePolicy(s, p, w, r)
	}
}

func handlePolicy(s *store.SourceStore, p store.Policy, w http.ResponseWriter, r *http.Request) {
	if err := s.AppendPolicy(p); err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func writeError(w http.ResponseWriter, err error, code int) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	})
}
