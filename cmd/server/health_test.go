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

func runPinger(t *testing.T, status int) error {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/healthz", r.URL.Path)
		w.WriteHeader(status)
	}))
	defer srv.Close()

	cmd := &cli.Command{
		Name: "ping",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "server-addr"},
			&cli.StringFlag{Name: "server-cert"},
		},
		Action: pinger,
	}
	return cmd.Run(t.Context(), []string{"ping", "--server-addr", strings.TrimPrefix(srv.URL, "http://")})
}

func TestPinger(t *testing.T) {
	for _, status := range []int{http.StatusOK, http.StatusNoContent} {
		assert.NoError(t, runPinger(t, status), "status %d", status)
	}
	for _, status := range []int{
		http.StatusMovedPermanently,
		http.StatusNotFound,
		http.StatusInternalServerError,
		http.StatusServiceUnavailable,
	} {
		assert.Error(t, runPinger(t, status), "status %d", status)
	}
}
