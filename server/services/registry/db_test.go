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

package registry_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/registry"
	mocks_store "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestRegistryListPipeline(t *testing.T) {
	globalRegistry := &model.Registry{ID: 1, Address: "docker.io", Username: "global"}
	orgRegistry := &model.Registry{ID: 2, OrgID: 1, Address: "docker.io", Username: "org"}
	repoRegistry := &model.Registry{ID: 3, RepoID: 1, Address: "docker.io", Username: "repo"}

	mockStore := mocks_store.NewStore(t)

	mockStore.On("RegistryList", mock.Anything, true, mock.Anything).Once().Return([]*model.Registry{
		globalRegistry,
		orgRegistry,
		repoRegistry,
	}, nil)

	r, err := registry.NewDB(mockStore).RegistryListPipeline(&model.Repo{}, &model.Pipeline{})
	assert.NoError(t, err)
	assert.Len(t, r, 1)
	assert.Equal(t, "repo", r[0].Username)

	mockStore.On("RegistryList", mock.Anything, true, mock.Anything).Once().Return([]*model.Registry{
		globalRegistry,
		orgRegistry,
	}, nil)

	r, err = registry.NewDB(mockStore).RegistryListPipeline(&model.Repo{}, &model.Pipeline{})
	assert.NoError(t, err)
	assert.Len(t, r, 1)
	assert.Equal(t, "org", r[0].Username)

	mockStore.On("RegistryList", mock.Anything, true, mock.Anything).Once().Return([]*model.Registry{
		globalRegistry,
	}, nil)

	r, err = registry.NewDB(mockStore).RegistryListPipeline(&model.Repo{}, &model.Pipeline{})
	assert.NoError(t, err)
	assert.Len(t, r, 1)
	assert.Equal(t, "global", r[0].Username)
}
