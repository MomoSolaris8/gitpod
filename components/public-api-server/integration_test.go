// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package main

import (
	"context"
	"github.com/gitpod-io/gitpod/common-go/baseserver"
	v1 "github.com/gitpod-io/gitpod/public-api/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"testing"
)

func TestPublicAPIServer_v1(t *testing.T) {
	t.SkipNow()
	ctx := context.Background()
	srv := baseserver.NewForTests(t)

	require.NoError(t, register(srv))
	baseserver.StartServerForTests(t, srv)

	addr := "api.mp-papi-caddy-grpc.staging.gitpod-dev.com:443"
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	workspaceClient := v1.NewWorkspacesServiceClient(conn)

	_, err = workspaceClient.GetWorkspace(ctx, &v1.GetWorkspaceRequest{})
	requireErrorStatusCode(t, codes.Unimplemented, err)
}

func TestPublicAPIServer_v1_PrebuildService(t *testing.T) {
	t.SkipNow()
	ctx := context.Background()
	srv := baseserver.NewForTests(t)
	require.NoError(t, register(srv))

	baseserver.StartServerForTests(t, srv)

	conn, err := grpc.Dial(srv.GRPCAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	prebuildClient := v1.NewPrebuildsServiceClient(conn)

	_, err = prebuildClient.GetPrebuild(ctx, &v1.GetPrebuildRequest{})
	requireErrorStatusCode(t, codes.Unimplemented, err)

	_, err = prebuildClient.GetRunningPrebuild(ctx, &v1.GetRunningPrebuildRequest{})
	requireErrorStatusCode(t, codes.Unimplemented, err)

	listenToStatusStream, err := prebuildClient.ListenToPrebuildStatus(ctx, &v1.ListenToPrebuildStatusRequest{})
	require.NoError(t, err)
	_, err = listenToStatusStream.Recv()
	requireErrorStatusCode(t, codes.Unimplemented, err)

	listenToLogsStream, err := prebuildClient.ListenToPrebuildLogs(ctx, &v1.ListenToPrebuildLogsRequest{})
	require.NoError(t, err)
	_, err = listenToLogsStream.Recv()
	requireErrorStatusCode(t, codes.Unimplemented, err)
}

func requireErrorStatusCode(t *testing.T, expected codes.Code, err error) {
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equalf(t, expected, st.Code(), "expected: %s but got: %s", expected.String(), st.String())
}