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

package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type fakeServerStream struct {
	grpc.ServerStream
	ctx  context.Context
	sent []any
}

func (s *fakeServerStream) Context() context.Context {
	return s.ctx
}

func (s *fakeServerStream) SendMsg(m any) error {
	s.sent = append(s.sent, m)
	return nil
}

func TestStreamInterceptor(t *testing.T) {
	jwtManager := NewJWTManager("secret")
	authorizer := NewAuthorizer(jwtManager)
	info := &grpc.StreamServerInfo{FullMethod: "/proto.Woodpecker/Next"}

	t.Run("passes authorized context and forwards stream calls", func(t *testing.T) {
		token, err := jwtManager.Generate(42)
		require.NoError(t, err)
		inner := &fakeServerStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.Pairs("token", token))}

		called := false
		err = authorizer.StreamInterceptor(nil, inner, info, func(_ any, stream grpc.ServerStream) error {
			called = true
			md, ok := metadata.FromIncomingContext(stream.Context())
			require.True(t, ok)
			assert.Equal(t, []string{"42"}, md.Get("agent_id"))
			return stream.SendMsg("hello")
		})

		require.NoError(t, err)
		assert.True(t, called)
		assert.Equal(t, []any{"hello"}, inner.sent)
	})

	t.Run("rejects missing token without calling handler", func(t *testing.T) {
		inner := &fakeServerStream{ctx: metadata.NewIncomingContext(context.Background(), metadata.MD{})}

		err := authorizer.StreamInterceptor(nil, inner, info, func(any, grpc.ServerStream) error {
			t.Fatal("handler must not be called")
			return nil
		})

		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}
