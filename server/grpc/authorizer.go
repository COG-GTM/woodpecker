// Copyright 2023 Woodpecker Authors
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

package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type wrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrapper) Context() context.Context {
	return w.ctx
}

type Authorizer struct {
	jwtManager *JWTManager
}

func NewAuthorizer(jwtManager *JWTManager) *Authorizer {
	return &Authorizer{jwtManager: jwtManager}
}

func (a *Authorizer) StreamInterceptor(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	newCtx, err := a.authorize(stream.Context(), info.FullMethod)
	if err != nil {
		return err
	}

	return handler(srv, &wrapper{ServerStream: stream, ctx: newCtx})
}

func (a *Authorizer) UnaryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	newCtx, err := a.authorize(ctx, info.FullMethod)
	if err != nil {
		return nil, err
	}
	return handler(newCtx, req)
}

func (a *Authorizer) authorize(ctx context.Context, fullMethod string) (context.Context, error) {
	// bypass auth for token endpoint
	if fullMethod == "/proto.WoodpeckerAuth/Auth" {
		return ctx, nil
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["token"]
	if len(values) == 0 {
		return ctx, status.Errorf(codes.Unauthenticated, "token is not provided")
	}

	accessToken := values[0]
	claims, err := a.jwtManager.Verify(accessToken)
	if err != nil {
		return ctx, status.Errorf(codes.Unauthenticated, "access token is invalid: %v", err)
	}

	md.Append("agent_id", fmt.Sprintf("%d", claims.AgentID))

	return metadata.NewIncomingContext(ctx, md), nil
}
