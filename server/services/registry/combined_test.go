// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package registry

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/registry/mocks"
)

type staticRegistries struct {
	ReadOnlyService
	list []*model.Registry
	err  error
}

func (s *staticRegistries) GlobalRegistryList(*model.ListOptions) ([]*model.Registry, error) {
	return s.list, s.err
}

func addresses(regs []*model.Registry) []string {
	out := make([]string, 0, len(regs))
	for _, r := range regs {
		out = append(out, r.Address)
	}
	return out
}

func TestCombinedPrefersDatabaseRegistries(t *testing.T) {
	repo := &model.Repo{ID: 1}
	pipeline := &model.Pipeline{ID: 2}

	dbRepo := &model.Registry{Address: "docker.io", Username: "db-repo"}
	dbGlobal := &model.Registry{Address: "ghcr.io", Username: "db-global"}

	db := mocks.NewService(t)
	db.On("RegistryListPipeline", repo, pipeline).Return([]*model.Registry{dbRepo}, nil)
	db.On("GlobalRegistryList", mock.Anything).Return([]*model.Registry{dbGlobal}, nil)

	extra := &staticRegistries{list: []*model.Registry{
		{Address: "docker.io", Username: "extra"},
		{Address: "quay.io", Username: "extra"},
		{Address: "quay.io", Username: "extra-dup"},
		{Address: "ghcr.io", Username: "extra"},
	}}

	c := NewCombined(db, extra)

	got, err := c.RegistryListPipeline(repo, pipeline)
	assert.NoError(t, err)
	assert.Equal(t, []string{"quay.io", "ghcr.io", "docker.io"}, addresses(got))
	assert.Equal(t, "extra", got[0].Username)
	assert.Equal(t, "extra", got[1].Username)
	assert.Equal(t, "db-repo", got[2].Username)

	got, err = c.GlobalRegistryList(&model.ListOptions{All: true})
	assert.NoError(t, err)
	assert.Equal(t, []string{"docker.io", "quay.io", "ghcr.io"}, addresses(got))
	assert.Equal(t, "extra", got[0].Username)
	assert.Equal(t, "db-global", got[2].Username)

	got, err = c.GlobalRegistryList(&model.ListOptions{Page: 2, PerPage: 2})
	assert.NoError(t, err)
	assert.Equal(t, []string{"ghcr.io"}, addresses(got))
}

func TestCombinedPropagatesSourceErrors(t *testing.T) {
	repo := &model.Repo{ID: 1}
	pipeline := &model.Pipeline{ID: 2}
	boom := errors.New("boom")

	db := mocks.NewService(t)
	db.On("RegistryListPipeline", repo, pipeline).Return([]*model.Registry{}, nil)
	db.On("GlobalRegistryList", mock.Anything).Return([]*model.Registry{}, nil)

	c := NewCombined(db, &staticRegistries{err: boom})

	_, err := c.RegistryListPipeline(repo, pipeline)
	assert.ErrorIs(t, err, boom)

	_, err = c.GlobalRegistryList(&model.ListOptions{All: true})
	assert.ErrorIs(t, err, boom)
}
