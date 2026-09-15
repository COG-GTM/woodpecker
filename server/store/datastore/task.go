// Copyright 2021 Woodpecker Authors
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
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func (s storage) TaskList() ([]*model.Task, error) {
	tasks := make([]*model.Task, 0, perPage)
	return tasks, s.engine.Find(&tasks)
}

func (s storage) TaskInsert(task *model.Task) error {
	// only Insert set auto created ID back to object
	_, err := s.engine.Insert(task)
	return err
}

func (s storage) TaskDelete(id string) error {
	return wrapDelete(s.engine.Where("id = ?", id).Delete(new(model.Task)))
}

func (s storage) TaskInsertAtOnce(tasks []*model.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	// insert in chunks to stay below the SQL placeholder limit
	const chunkSize = 100
	for start := 0; start < len(tasks); start += chunkSize {
		end := start + chunkSize
		if end > len(tasks) {
			end = len(tasks)
		}
		if _, err := sess.Insert(tasks[start:end]); err != nil {
			return err
		}
	}

	return sess.Commit()
}

func (s storage) TaskDeleteAtOnce(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := s.engine.In("id", ids).Delete(new(model.Task))
	return err
}
