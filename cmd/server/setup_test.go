// Copyright 2025 Woodpecker Authors
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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	mocks_store "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func newTestCommand(t *testing.T, args ...string) *cli.Command {
	t.Helper()
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "grpc-secret"},
		},
		Action: func(_ context.Context, _ *cli.Command) error { return nil },
	}
	require.NoError(t, cmd.Run(context.Background(), append([]string{"server"}, args...)))
	return cmd
}

func TestSetupGRPCSecret(t *testing.T) {
	t.Run("explicit value is returned as-is", func(t *testing.T) {
		store := mocks_store.NewStore(t)
		secret, err := setupGRPCSecret(newTestCommand(t, "--grpc-secret=strong-secret"), store)
		assert.NoError(t, err)
		assert.Equal(t, "strong-secret", secret)
	})

	t.Run("literal \"secret\" is rejected", func(t *testing.T) {
		store := mocks_store.NewStore(t)
		secret, err := setupGRPCSecret(newTestCommand(t, "--grpc-secret=secret"), store)
		assert.Error(t, err)
		assert.Empty(t, secret)
	})

	t.Run("generates and persists a secret when unset", func(t *testing.T) {
		store := mocks_store.NewStore(t)
		store.On("ServerConfigGet", grpcSecretID).Return("", types.RecordNotExist)
		store.On("ServerConfigSet", grpcSecretID, mock.Anything).Return(nil)

		secret, err := setupGRPCSecret(newTestCommand(t), store)
		assert.NoError(t, err)
		assert.NotEmpty(t, secret)
		store.AssertCalled(t, "ServerConfigSet", grpcSecretID, secret)
	})

	t.Run("returns stored secret when one exists", func(t *testing.T) {
		store := mocks_store.NewStore(t)
		store.On("ServerConfigGet", grpcSecretID).Return("stored-secret", nil)

		secret, err := setupGRPCSecret(newTestCommand(t), store)
		assert.NoError(t, err)
		assert.Equal(t, "stored-secret", secret)
	})
}
