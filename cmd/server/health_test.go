// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v3"
)

func runPinger(t *testing.T, handler http.Handler, serverHost string) error {
	t.Helper()
	srv := httptest.NewServer(handler)
	defer srv.Close()

	cmd := &cli.Command{
		Name: "ping",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "server-addr"},
			&cli.StringFlag{Name: "server-cert"},
			&cli.StringFlag{Name: "server-host"},
		},
		Action: pinger,
	}
	return cmd.Run(t.Context(), []string{
		"ping",
		"--server-addr", strings.TrimPrefix(srv.URL, "http://"),
		"--server-host", serverHost,
	})
}

// healthServer serves status on healthPath and 200 (like the web UI fallback) everywhere else.
func healthServer(healthPath string, status int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == healthPath {
			w.WriteHeader(status)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestPingerStatus(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusNoContent} {
		assert.NoError(t, runPinger(t, healthServer("/healthz", status), "http://example.com"), "status %d", status)
	}
	for _, status := range []int{
		http.StatusMovedPermanently,
		http.StatusNotFound,
		http.StatusInternalServerError,
		http.StatusServiceUnavailable,
	} {
		assert.Error(t, runPinger(t, healthServer("/healthz", status), "http://example.com"), "status %d", status)
	}
}

func TestPingerRootPath(t *testing.T) {
	for _, host := range []string{"https://example.com/ci", "https://example.com/ci/"} {
		assert.NoError(t, runPinger(t, healthServer("/ci/healthz", http.StatusNoContent), host), host)
		assert.Error(t, runPinger(t, healthServer("/ci/healthz", http.StatusInternalServerError), host), host)
	}
}

func TestPingerDoesNotFollowRedirects(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	assert.Error(t, runPinger(t, handler, "http://example.com"))
}
