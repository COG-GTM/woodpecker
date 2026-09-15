// Copyright 2018 Drone.IO Inc.
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

package datastore

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestTaskList(t *testing.T) {
	store, closer := newTestStore(t, new(model.Task))
	defer closer()

	assert.NoError(t, store.TaskInsert(&model.Task{
		ID:        "some_random_id",
		Data:      []byte("foo"),
		Labels:    map[string]string{"foo": "bar"},
		DepStatus: map[string]model.StatusValue{"test": "dep"},
	}))

	list, err := store.TaskList()
	assert.NoError(t, err)
	assert.Len(t, list, 1, "Expected one task in list")
	assert.Equal(t, "some_random_id", list[0].ID)
	assert.Equal(t, "foo", string(list[0].Data))
	assert.EqualValues(t, map[string]model.StatusValue{"test": "dep"}, list[0].DepStatus)

	assert.NoError(t, store.TaskDelete("some_random_id"))

	list, err = store.TaskList()
	assert.NoError(t, err)
	assert.Len(t, list, 0, "Want empty task list after delete")
}

func TestTaskInsertAndDeleteAtOnce(t *testing.T) {
	store, closer := newTestStore(t, new(model.Task))
	defer closer()

	tasks := []*model.Task{
		{
			ID:           "task_1",
			Data:         []byte("foo"),
			Labels:       map[string]string{"foo": "bar"},
			Dependencies: []string{"dep_a"},
		},
		{
			ID:           "task_2",
			Data:         []byte("bar"),
			Labels:       map[string]string{"baz": "qux"},
			Dependencies: []string{"dep_b"},
		},
		{
			ID:           "task_3",
			Data:         []byte("baz"),
			Labels:       map[string]string{"a": "b"},
			Dependencies: []string{"dep_c"},
		},
	}
	assert.NoError(t, store.TaskInsertAtOnce(tasks))

	list, err := store.TaskList()
	assert.NoError(t, err)
	assert.Len(t, list, 3, "Expected three tasks in list")
	for i, task := range tasks {
		assert.Equal(t, task.ID, list[i].ID)
		assert.EqualValues(t, task.Labels, list[i].Labels)
		assert.EqualValues(t, task.Dependencies, list[i].Dependencies)
	}

	assert.NoError(t, store.TaskDeleteAtOnce([]string{"task_1", "task_3"}))

	list, err = store.TaskList()
	assert.NoError(t, err)
	assert.Len(t, list, 1, "Want one task left in list")
	assert.Equal(t, "task_2", list[0].ID)

	// deleting a missing id must not error
	assert.NoError(t, store.TaskDeleteAtOnce([]string{"task_2", "missing_id"}))
	// empty slice is a no-op
	assert.NoError(t, store.TaskDeleteAtOnce([]string{}))
	assert.NoError(t, store.TaskInsertAtOnce(nil))
}
