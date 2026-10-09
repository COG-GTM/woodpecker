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

package common_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/common"
)

func Test_Netrc(t *testing.T) {
	host, err := common.ExtractHostFromCloneURL("https://git.example.com/foo/bar.git")
	assert.NoError(t, err)
	assert.Equal(t, "git.example.com", host)
}

func Test_FixMalformedAvatar(t *testing.T) {
	urls := []struct {
		Before string
		After  string
	}{
		{
			"http://gitea.golang.org///1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
			"//1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
		},
		{
			"//1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
			"//1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
		},
		{
			"http://gitea.golang.org/avatars/1",
			"http://gitea.golang.org/avatars/1",
		},
		{
			"http://gitea.golang.org//avatars/1",
			"http://gitea.golang.org/avatars/1",
		},
	}

	for _, url := range urls {
		got := common.FixMalformedAvatar(url.Before)
		assert.Equal(t, url.After, got)
	}
}

func Test_ExpandAvatar(t *testing.T) {
	urls := []struct {
		Before string
		After  string
	}{
		{
			"/avatars/1",
			"http://gitea.io/avatars/1",
		},
		{
			"//1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
			"http://1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87",
		},
		{
			"/gitea/avatars/2",
			"http://gitea.io/gitea/avatars/2",
		},
	}

	repo := "http://gitea.io/foo/bar"
	for _, url := range urls {
		got := common.ExpandAvatar(repo, url.Before)
		assert.Equal(t, url.After, got)
	}
}
