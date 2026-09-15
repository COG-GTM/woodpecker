// Copyright 2022 Woodpecker Authors
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

package grpc

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	mocks_store "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestRegisterAgent(t *testing.T) {
	t.Run("When existing agent Name is empty it should update Name with hostname from metadata", func(t *testing.T) {
		store := mocks_store.NewStore(t)
		storeAgent := new(model.Agent)
		storeAgent.ID = 1337
		updatedAgent := model.Agent{
			ID:          1337,
			Created:     0,
			Updated:     0,
			Name:        "hostname",
			OwnerID:     0,
			Token:       "",
			LastContact: 0,
			Platform:    "platform",
			Backend:     "backend",
			Capacity:    2,
			Version:     "version",
			NoSchedule:  false,
		}

		store.On("AgentFind", int64(1337)).Once().Return(storeAgent, nil)
		store.On("AgentUpdate", &updatedAgent).Once().Return(nil)
		grpc := RPC{
			store: store,
		}
		ctx := metadata.NewIncomingContext(
			t.Context(),
			metadata.Pairs("hostname", "hostname", "agent_id", "1337"),
		)
		agentID, err := grpc.RegisterAgent(ctx, rpc.AgentInfo{
			Version:  "version",
			Platform: "platform",
			Backend:  "backend",
			Capacity: 2,
		})
		if !assert.NoError(t, err) {
			return
		}

		assert.EqualValues(t, 1337, agentID)
	})

	t.Run("When existing agent hostname is present it should not update the hostname", func(t *testing.T) {
		store := mocks_store.NewStore(t)
		storeAgent := new(model.Agent)
		storeAgent.ID = 1337
		storeAgent.Name = "originalHostname"
		updatedAgent := model.Agent{
			ID:          1337,
			Created:     0,
			Updated:     0,
			Name:        "originalHostname",
			OwnerID:     0,
			Token:       "",
			LastContact: 0,
			Platform:    "platform",
			Backend:     "backend",
			Capacity:    2,
			Version:     "version",
			NoSchedule:  false,
		}

		store.On("AgentFind", int64(1337)).Once().Return(storeAgent, nil)
		store.On("AgentUpdate", &updatedAgent).Once().Return(nil)
		grpc := RPC{
			store: store,
		}
		ctx := metadata.NewIncomingContext(
			t.Context(),
			metadata.Pairs("hostname", "newHostname", "agent_id", "1337"),
		)
		agentID, err := grpc.RegisterAgent(ctx, rpc.AgentInfo{
			Version:  "version",
			Platform: "platform",
			Backend:  "backend",
			Capacity: 2,
		})
		if !assert.NoError(t, err) {
			return
		}

		assert.EqualValues(t, 1337, agentID)
	})
}

func TestUpdateAgentLastWork(t *testing.T) {
	t.Run("When last work was never updated it should update last work timestamp", func(t *testing.T) {
		agent := model.Agent{
			LastWork: 0,
		}
		store := mocks_store.NewStore(t)
		rpc := RPC{
			store: store,
		}
		store.On("AgentUpdate", mock.Anything).Once().Return(nil)

		err := rpc.updateAgentLastWork(&agent)
		assert.NoError(t, err)

		assert.NotZero(t, agent.LastWork)
	})

	t.Run("When last work was updated over a minute ago it should update last work timestamp", func(t *testing.T) {
		lastWork := time.Now().Add(-time.Hour).Unix()
		agent := model.Agent{
			LastWork: lastWork,
		}
		store := mocks_store.NewStore(t)
		rpc := RPC{
			store: store,
		}
		store.On("AgentUpdate", mock.Anything).Once().Return(nil)

		err := rpc.updateAgentLastWork(&agent)
		assert.NoError(t, err)

		assert.NotEqual(t, lastWork, agent.LastWork)
	})

	t.Run("When last work was updated in the last minute it should not update last work timestamp again", func(t *testing.T) {
		lastWork := time.Now().Add(-time.Second * 30).Unix()
		agent := model.Agent{
			LastWork: lastWork,
		}
		rpc := RPC{}

		err := rpc.updateAgentLastWork(&agent)
		assert.NoError(t, err)

		assert.Equal(t, lastWork, agent.LastWork)
	})
}

func TestGetLogStepContext(t *testing.T) {
	t.Run("It should resolve step, agent, pipeline and repo once and serve later calls from cache", func(t *testing.T) {
		store := mocks_store.NewStore(t)
		step := &model.Step{ID: 7, UUID: "step-uuid", PipelineID: 3}
		agent := &model.Agent{ID: 1337, OrgID: 42, LastWork: time.Now().Unix()}
		pipe := &model.Pipeline{ID: 3, RepoID: 5}
		repo := &model.Repo{ID: 5, OrgID: 42}

		store.On("StepByUUID", "step-uuid").Once().Return(step, nil)
		store.On("AgentFind", int64(1337)).Once().Return(agent, nil)
		store.On("GetPipeline", int64(3)).Once().Return(pipe, nil)
		store.On("GetRepo", int64(5)).Once().Return(repo, nil)

		rpc := RPC{store: store, logStepCache: new(sync.Map)}
		for range 3 {
			stepCtx, err := rpc.getLogStepContext(t.Context(), 1337, "step-uuid")
			assert.NoError(t, err)
			assert.Equal(t, step, stepCtx.step)
			assert.NoError(t, rpc.updateCachedAgentLastWork(stepCtx))
		}

		rpc.invalidateLogStepCache([]*model.Step{step})
		_, ok := rpc.logStepCache.Load("step-uuid")
		assert.False(t, ok)
	})

	t.Run("It should not cache a denied permission", func(t *testing.T) {
		store := mocks_store.NewStore(t)
		step := &model.Step{ID: 7, UUID: "step-uuid", PipelineID: 3}
		agent := &model.Agent{ID: 1337, OrgID: 1}
		pipe := &model.Pipeline{ID: 3, RepoID: 5}
		repo := &model.Repo{ID: 5, OrgID: 42}

		store.On("StepByUUID", "step-uuid").Return(step, nil)
		store.On("AgentFind", int64(1337)).Return(agent, nil)
		store.On("GetPipeline", int64(3)).Return(pipe, nil)
		store.On("GetRepo", int64(5)).Return(repo, nil)

		rpc := RPC{store: store, logStepCache: new(sync.Map)}
		_, err := rpc.getLogStepContext(t.Context(), 1337, "step-uuid")
		assert.Error(t, err)
		_, ok := rpc.logStepCache.Load("step-uuid")
		assert.False(t, ok)
	})
}
