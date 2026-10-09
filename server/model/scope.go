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

// Scoped is implemented by items that can be defined on repository, organization or global level.
type Scoped interface {
	IsRepository() bool
	IsOrganization() bool
	IsGlobal() bool
}

// UniqueByScopePriority returns the items with a unique key.
// Priority order in case of duplicate keys are repository, user/organization, global;
// within the same scope the first item wins.
func UniqueByScopePriority[T Scoped](items []T, key func(T) string) []T {
	result := make([]T, 0, len(items))
	uniq := make(map[string]struct{})
	for _, condition := range []struct {
		IsRepository   bool
		IsOrganization bool
		IsGlobal       bool
	}{
		{IsRepository: true},
		{IsOrganization: true},
		{IsGlobal: true},
	} {
		for _, item := range items {
			if item.IsRepository() != condition.IsRepository || item.IsOrganization() != condition.IsOrganization || item.IsGlobal() != condition.IsGlobal {
				continue
			}
			k := key(item)
			if _, ok := uniq[k]; ok {
				continue
			}
			uniq[k] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
