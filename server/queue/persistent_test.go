// Copyright 2025 Woodpecker Authors
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

//go:build test
// +build test

package queue

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	mocks_store "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestPersistentQueuePushAtOnce(t *testing.T) {
	ctx := context.Background()
	store := new(mocks_store.Store)
	store.On("TaskList").Return([]*model.Task{}, nil)
	q := WithTaskStore(ctx, NewMemoryQueue(ctx), store)

	tasks := []*model.Task{
		{ID: "1", Data: []byte("{}")},
		{ID: "2", Data: []byte("{}")},
		{ID: "3", Data: []byte("{}")},
	}

	store.On("TaskInsertAtOnce", tasks).Return(nil).Once()
	assert.NoError(t, q.PushAtOnce(ctx, tasks))

	store.AssertExpectations(t)
	store.AssertNotCalled(t, "TaskDeleteAtOnce", mock.Anything)
}

func TestPersistentQueuePushAtOnceStoreError(t *testing.T) {
	ctx := context.Background()
	store := new(mocks_store.Store)
	store.On("TaskList").Return([]*model.Task{}, nil)
	q := WithTaskStore(ctx, NewMemoryQueue(ctx), store)

	tasks := []*model.Task{
		{ID: "1", Data: []byte("{}")},
		{ID: "2", Data: []byte("{}")},
		{ID: "3", Data: []byte("{}")},
	}

	storeErr := errors.New("insert failed")
	store.On("TaskInsertAtOnce", tasks).Return(storeErr).Once()
	assert.ErrorIs(t, q.PushAtOnce(ctx, tasks), storeErr)

	pollCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()
	task, err := q.Poll(pollCtx, 1, filterFnTrue)
	assert.Error(t, err)
	assert.Nil(t, task, "Want empty queue after failed insert")

	store.AssertExpectations(t)
}
