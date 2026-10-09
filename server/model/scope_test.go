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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniqueByScopePriority(t *testing.T) {
	global := &Registry{ID: 1, Address: "docker.io"}
	org := &Registry{ID: 2, OrgID: 1, Address: "docker.io"}
	repo := &Registry{ID: 3, RepoID: 1, Address: "docker.io"}
	repoDup := &Registry{ID: 4, RepoID: 1, Address: "docker.io"}
	globalOther := &Registry{ID: 5, Address: "ghcr.io"}
	orgOther := &Registry{ID: 6, OrgID: 1, Address: "quay.io"}
	invalidScope := &Registry{ID: 7, OrgID: 1, RepoID: 1, Address: "invalid.io"}

	byAddress := func(r *Registry) string { return r.Address }

	assert.Empty(t, UniqueByScopePriority(nil, byAddress))

	assert.Equal(t,
		[]*Registry{repo, orgOther, globalOther},
		UniqueByScopePriority([]*Registry{globalOther, global, org, orgOther, repo, repoDup, invalidScope}, byAddress),
	)
	assert.Equal(t,
		[]*Registry{org},
		UniqueByScopePriority([]*Registry{global, org}, byAddress),
	)
	assert.Equal(t,
		[]*Registry{global},
		UniqueByScopePriority([]*Registry{global}, byAddress),
	)

	secrets := UniqueByScopePriority([]*Secret{
		{ID: 1, Name: "token", Value: "global"},
		{ID: 2, RepoID: 1, Name: "token", Value: "repo"},
		{ID: 3, OrgID: 1, Name: "other", Value: "org"},
	}, func(s *Secret) string { return s.Name })
	assert.Len(t, secrets, 2)
	assert.Equal(t, "repo", secrets[0].Value)
	assert.Equal(t, "org", secrets[1].Value)
}
