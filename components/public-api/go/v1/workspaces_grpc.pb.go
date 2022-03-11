// Copyright (c) 2022 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// WorkspacesServiceClient is the client API for WorkspacesService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WorkspacesServiceClient interface {
	// ListWorkspaces enumerates all workspaces belonging to the authenticated user.
	ListWorkspaces(ctx context.Context, in *ListWorkspacesRequest, opts ...grpc.CallOption) (*ListWorkspacesResponse, error)
	// GetWorkspace returns a single workspace.
	GetWorkspace(ctx context.Context, in *GetWorkspaceRequest, opts ...grpc.CallOption) (*GetWorkspaceResponse, error)
	// CreateAndStartWorkspace creates a new workspace and starts it.
	CreateAndStartWorkspace(ctx context.Context, in *CreateAndStartWorkspaceRequest, opts ...grpc.CallOption) (*CreateAndStartWorkspaceResponse, error)
	// StartWorkspace starts an existing workspace.
	StartWorkspace(ctx context.Context, in *StartWorkspaceRequest, opts ...grpc.CallOption) (*StartWorkspaceResponse, error)
	// GetRunningWorkspaceInstance returns the currently active instance of a workspace.
	// Errors:
	//   FAILED_PRECONDITION: if a workspace does not a currently active instance
	//
	GetActiveWorkspaceInstance(ctx context.Context, in *GetActiveWorkspaceInstanceRequest, opts ...grpc.CallOption) (*GetActiveWorkspaceInstanceResponse, error)
	// GetWorkspaceInstanceOwnerToken returns the owner token of a workspace instance.
	// Note: the owner token is not part of the workspace instance status so that we can scope its access on the
	//       API function level.
	GetWorkspaceInstanceOwnerToken(ctx context.Context, in *GetWorkspaceInstanceOwnerTokenRequest, opts ...grpc.CallOption) (*GetWorkspaceInstanceOwnerTokenResponse, error)
	// ListenToWorkspaceInstance listens to workspace instance updates.
	ListenToWorkspaceInstance(ctx context.Context, in *ListenToWorkspaceInstanceRequest, opts ...grpc.CallOption) (WorkspacesService_ListenToWorkspaceInstanceClient, error)
	// ListenToImageBuildLogs streams (currently or previously) running workspace image build logs
	ListenToImageBuildLogs(ctx context.Context, in *ListenToImageBuildLogsRequest, opts ...grpc.CallOption) (WorkspacesService_ListenToImageBuildLogsClient, error)
	// StopWorkspace stops a running workspace (instance).
	// Errors:
	//   NOT_FOUND:           the workspace_id is unkown
	//   FAILED_PRECONDITION: if there's no running instance
	StopWorkspace(ctx context.Context, in *StopWorkspaceRequest, opts ...grpc.CallOption) (WorkspacesService_StopWorkspaceClient, error)
}

type workspacesServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewWorkspacesServiceClient(cc grpc.ClientConnInterface) WorkspacesServiceClient {
	return &workspacesServiceClient{cc}
}

func (c *workspacesServiceClient) ListWorkspaces(ctx context.Context, in *ListWorkspacesRequest, opts ...grpc.CallOption) (*ListWorkspacesResponse, error) {
	out := new(ListWorkspacesResponse)
	err := c.cc.Invoke(ctx, "/gitpod.v1.WorkspacesService/ListWorkspaces", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workspacesServiceClient) GetWorkspace(ctx context.Context, in *GetWorkspaceRequest, opts ...grpc.CallOption) (*GetWorkspaceResponse, error) {
	out := new(GetWorkspaceResponse)
	err := c.cc.Invoke(ctx, "/gitpod.v1.WorkspacesService/GetWorkspace", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workspacesServiceClient) CreateAndStartWorkspace(ctx context.Context, in *CreateAndStartWorkspaceRequest, opts ...grpc.CallOption) (*CreateAndStartWorkspaceResponse, error) {
	out := new(CreateAndStartWorkspaceResponse)
	err := c.cc.Invoke(ctx, "/gitpod.v1.WorkspacesService/CreateAndStartWorkspace", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workspacesServiceClient) StartWorkspace(ctx context.Context, in *StartWorkspaceRequest, opts ...grpc.CallOption) (*StartWorkspaceResponse, error) {
	out := new(StartWorkspaceResponse)
	err := c.cc.Invoke(ctx, "/gitpod.v1.WorkspacesService/StartWorkspace", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workspacesServiceClient) GetActiveWorkspaceInstance(ctx context.Context, in *GetActiveWorkspaceInstanceRequest, opts ...grpc.CallOption) (*GetActiveWorkspaceInstanceResponse, error) {
	out := new(GetActiveWorkspaceInstanceResponse)
	err := c.cc.Invoke(ctx, "/gitpod.v1.WorkspacesService/GetActiveWorkspaceInstance", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workspacesServiceClient) GetWorkspaceInstanceOwnerToken(ctx context.Context, in *GetWorkspaceInstanceOwnerTokenRequest, opts ...grpc.CallOption) (*GetWorkspaceInstanceOwnerTokenResponse, error) {
	out := new(GetWorkspaceInstanceOwnerTokenResponse)
	err := c.cc.Invoke(ctx, "/gitpod.v1.WorkspacesService/GetWorkspaceInstanceOwnerToken", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workspacesServiceClient) ListenToWorkspaceInstance(ctx context.Context, in *ListenToWorkspaceInstanceRequest, opts ...grpc.CallOption) (WorkspacesService_ListenToWorkspaceInstanceClient, error) {
	stream, err := c.cc.NewStream(ctx, &WorkspacesService_ServiceDesc.Streams[0], "/gitpod.v1.WorkspacesService/ListenToWorkspaceInstance", opts...)
	if err != nil {
		return nil, err
	}
	x := &workspacesServiceListenToWorkspaceInstanceClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type WorkspacesService_ListenToWorkspaceInstanceClient interface {
	Recv() (*ListenToWorkspaceInstanceResponse, error)
	grpc.ClientStream
}

type workspacesServiceListenToWorkspaceInstanceClient struct {
	grpc.ClientStream
}

func (x *workspacesServiceListenToWorkspaceInstanceClient) Recv() (*ListenToWorkspaceInstanceResponse, error) {
	m := new(ListenToWorkspaceInstanceResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *workspacesServiceClient) ListenToImageBuildLogs(ctx context.Context, in *ListenToImageBuildLogsRequest, opts ...grpc.CallOption) (WorkspacesService_ListenToImageBuildLogsClient, error) {
	stream, err := c.cc.NewStream(ctx, &WorkspacesService_ServiceDesc.Streams[1], "/gitpod.v1.WorkspacesService/ListenToImageBuildLogs", opts...)
	if err != nil {
		return nil, err
	}
	x := &workspacesServiceListenToImageBuildLogsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type WorkspacesService_ListenToImageBuildLogsClient interface {
	Recv() (*ListenToImageBuildLogsResponse, error)
	grpc.ClientStream
}

type workspacesServiceListenToImageBuildLogsClient struct {
	grpc.ClientStream
}

func (x *workspacesServiceListenToImageBuildLogsClient) Recv() (*ListenToImageBuildLogsResponse, error) {
	m := new(ListenToImageBuildLogsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *workspacesServiceClient) StopWorkspace(ctx context.Context, in *StopWorkspaceRequest, opts ...grpc.CallOption) (WorkspacesService_StopWorkspaceClient, error) {
	stream, err := c.cc.NewStream(ctx, &WorkspacesService_ServiceDesc.Streams[2], "/gitpod.v1.WorkspacesService/StopWorkspace", opts...)
	if err != nil {
		return nil, err
	}
	x := &workspacesServiceStopWorkspaceClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type WorkspacesService_StopWorkspaceClient interface {
	Recv() (*StopWorkspaceResponse, error)
	grpc.ClientStream
}

type workspacesServiceStopWorkspaceClient struct {
	grpc.ClientStream
}

func (x *workspacesServiceStopWorkspaceClient) Recv() (*StopWorkspaceResponse, error) {
	m := new(StopWorkspaceResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// WorkspacesServiceServer is the server API for WorkspacesService service.
// All implementations must embed UnimplementedWorkspacesServiceServer
// for forward compatibility
type WorkspacesServiceServer interface {
	// ListWorkspaces enumerates all workspaces belonging to the authenticated user.
	ListWorkspaces(context.Context, *ListWorkspacesRequest) (*ListWorkspacesResponse, error)
	// GetWorkspace returns a single workspace.
	GetWorkspace(context.Context, *GetWorkspaceRequest) (*GetWorkspaceResponse, error)
	// CreateAndStartWorkspace creates a new workspace and starts it.
	CreateAndStartWorkspace(context.Context, *CreateAndStartWorkspaceRequest) (*CreateAndStartWorkspaceResponse, error)
	// StartWorkspace starts an existing workspace.
	StartWorkspace(context.Context, *StartWorkspaceRequest) (*StartWorkspaceResponse, error)
	// GetRunningWorkspaceInstance returns the currently active instance of a workspace.
	// Errors:
	//   FAILED_PRECONDITION: if a workspace does not a currently active instance
	//
	GetActiveWorkspaceInstance(context.Context, *GetActiveWorkspaceInstanceRequest) (*GetActiveWorkspaceInstanceResponse, error)
	// GetWorkspaceInstanceOwnerToken returns the owner token of a workspace instance.
	// Note: the owner token is not part of the workspace instance status so that we can scope its access on the
	//       API function level.
	GetWorkspaceInstanceOwnerToken(context.Context, *GetWorkspaceInstanceOwnerTokenRequest) (*GetWorkspaceInstanceOwnerTokenResponse, error)
	// ListenToWorkspaceInstance listens to workspace instance updates.
	ListenToWorkspaceInstance(*ListenToWorkspaceInstanceRequest, WorkspacesService_ListenToWorkspaceInstanceServer) error
	// ListenToImageBuildLogs streams (currently or previously) running workspace image build logs
	ListenToImageBuildLogs(*ListenToImageBuildLogsRequest, WorkspacesService_ListenToImageBuildLogsServer) error
	// StopWorkspace stops a running workspace (instance).
	// Errors:
	//   NOT_FOUND:           the workspace_id is unkown
	//   FAILED_PRECONDITION: if there's no running instance
	StopWorkspace(*StopWorkspaceRequest, WorkspacesService_StopWorkspaceServer) error
	mustEmbedUnimplementedWorkspacesServiceServer()
}

// UnimplementedWorkspacesServiceServer must be embedded to have forward compatible implementations.
type UnimplementedWorkspacesServiceServer struct {
}

func (UnimplementedWorkspacesServiceServer) ListWorkspaces(context.Context, *ListWorkspacesRequest) (*ListWorkspacesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListWorkspaces not implemented")
}
func (UnimplementedWorkspacesServiceServer) GetWorkspace(context.Context, *GetWorkspaceRequest) (*GetWorkspaceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWorkspace not implemented")
}
func (UnimplementedWorkspacesServiceServer) CreateAndStartWorkspace(context.Context, *CreateAndStartWorkspaceRequest) (*CreateAndStartWorkspaceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAndStartWorkspace not implemented")
}
func (UnimplementedWorkspacesServiceServer) StartWorkspace(context.Context, *StartWorkspaceRequest) (*StartWorkspaceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StartWorkspace not implemented")
}
func (UnimplementedWorkspacesServiceServer) GetActiveWorkspaceInstance(context.Context, *GetActiveWorkspaceInstanceRequest) (*GetActiveWorkspaceInstanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetActiveWorkspaceInstance not implemented")
}
func (UnimplementedWorkspacesServiceServer) GetWorkspaceInstanceOwnerToken(context.Context, *GetWorkspaceInstanceOwnerTokenRequest) (*GetWorkspaceInstanceOwnerTokenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWorkspaceInstanceOwnerToken not implemented")
}
func (UnimplementedWorkspacesServiceServer) ListenToWorkspaceInstance(*ListenToWorkspaceInstanceRequest, WorkspacesService_ListenToWorkspaceInstanceServer) error {
	return status.Errorf(codes.Unimplemented, "method ListenToWorkspaceInstance not implemented")
}
func (UnimplementedWorkspacesServiceServer) ListenToImageBuildLogs(*ListenToImageBuildLogsRequest, WorkspacesService_ListenToImageBuildLogsServer) error {
	return status.Errorf(codes.Unimplemented, "method ListenToImageBuildLogs not implemented")
}
func (UnimplementedWorkspacesServiceServer) StopWorkspace(*StopWorkspaceRequest, WorkspacesService_StopWorkspaceServer) error {
	return status.Errorf(codes.Unimplemented, "method StopWorkspace not implemented")
}
func (UnimplementedWorkspacesServiceServer) mustEmbedUnimplementedWorkspacesServiceServer() {}

// UnsafeWorkspacesServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WorkspacesServiceServer will
// result in compilation errors.
type UnsafeWorkspacesServiceServer interface {
	mustEmbedUnimplementedWorkspacesServiceServer()
}

func RegisterWorkspacesServiceServer(s grpc.ServiceRegistrar, srv WorkspacesServiceServer) {
	s.RegisterService(&WorkspacesService_ServiceDesc, srv)
}

func _WorkspacesService_ListWorkspaces_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListWorkspacesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkspacesServiceServer).ListWorkspaces(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitpod.v1.WorkspacesService/ListWorkspaces",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkspacesServiceServer).ListWorkspaces(ctx, req.(*ListWorkspacesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkspacesService_GetWorkspace_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetWorkspaceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkspacesServiceServer).GetWorkspace(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitpod.v1.WorkspacesService/GetWorkspace",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkspacesServiceServer).GetWorkspace(ctx, req.(*GetWorkspaceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkspacesService_CreateAndStartWorkspace_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateAndStartWorkspaceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkspacesServiceServer).CreateAndStartWorkspace(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitpod.v1.WorkspacesService/CreateAndStartWorkspace",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkspacesServiceServer).CreateAndStartWorkspace(ctx, req.(*CreateAndStartWorkspaceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkspacesService_StartWorkspace_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartWorkspaceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkspacesServiceServer).StartWorkspace(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitpod.v1.WorkspacesService/StartWorkspace",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkspacesServiceServer).StartWorkspace(ctx, req.(*StartWorkspaceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkspacesService_GetActiveWorkspaceInstance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetActiveWorkspaceInstanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkspacesServiceServer).GetActiveWorkspaceInstance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitpod.v1.WorkspacesService/GetActiveWorkspaceInstance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkspacesServiceServer).GetActiveWorkspaceInstance(ctx, req.(*GetActiveWorkspaceInstanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkspacesService_GetWorkspaceInstanceOwnerToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetWorkspaceInstanceOwnerTokenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkspacesServiceServer).GetWorkspaceInstanceOwnerToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/gitpod.v1.WorkspacesService/GetWorkspaceInstanceOwnerToken",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkspacesServiceServer).GetWorkspaceInstanceOwnerToken(ctx, req.(*GetWorkspaceInstanceOwnerTokenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkspacesService_ListenToWorkspaceInstance_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ListenToWorkspaceInstanceRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(WorkspacesServiceServer).ListenToWorkspaceInstance(m, &workspacesServiceListenToWorkspaceInstanceServer{stream})
}

type WorkspacesService_ListenToWorkspaceInstanceServer interface {
	Send(*ListenToWorkspaceInstanceResponse) error
	grpc.ServerStream
}

type workspacesServiceListenToWorkspaceInstanceServer struct {
	grpc.ServerStream
}

func (x *workspacesServiceListenToWorkspaceInstanceServer) Send(m *ListenToWorkspaceInstanceResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _WorkspacesService_ListenToImageBuildLogs_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ListenToImageBuildLogsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(WorkspacesServiceServer).ListenToImageBuildLogs(m, &workspacesServiceListenToImageBuildLogsServer{stream})
}

type WorkspacesService_ListenToImageBuildLogsServer interface {
	Send(*ListenToImageBuildLogsResponse) error
	grpc.ServerStream
}

type workspacesServiceListenToImageBuildLogsServer struct {
	grpc.ServerStream
}

func (x *workspacesServiceListenToImageBuildLogsServer) Send(m *ListenToImageBuildLogsResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _WorkspacesService_StopWorkspace_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(StopWorkspaceRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(WorkspacesServiceServer).StopWorkspace(m, &workspacesServiceStopWorkspaceServer{stream})
}

type WorkspacesService_StopWorkspaceServer interface {
	Send(*StopWorkspaceResponse) error
	grpc.ServerStream
}

type workspacesServiceStopWorkspaceServer struct {
	grpc.ServerStream
}

func (x *workspacesServiceStopWorkspaceServer) Send(m *StopWorkspaceResponse) error {
	return x.ServerStream.SendMsg(m)
}

// WorkspacesService_ServiceDesc is the grpc.ServiceDesc for WorkspacesService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var WorkspacesService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gitpod.v1.WorkspacesService",
	HandlerType: (*WorkspacesServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListWorkspaces",
			Handler:    _WorkspacesService_ListWorkspaces_Handler,
		},
		{
			MethodName: "GetWorkspace",
			Handler:    _WorkspacesService_GetWorkspace_Handler,
		},
		{
			MethodName: "CreateAndStartWorkspace",
			Handler:    _WorkspacesService_CreateAndStartWorkspace_Handler,
		},
		{
			MethodName: "StartWorkspace",
			Handler:    _WorkspacesService_StartWorkspace_Handler,
		},
		{
			MethodName: "GetActiveWorkspaceInstance",
			Handler:    _WorkspacesService_GetActiveWorkspaceInstance_Handler,
		},
		{
			MethodName: "GetWorkspaceInstanceOwnerToken",
			Handler:    _WorkspacesService_GetWorkspaceInstanceOwnerToken_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ListenToWorkspaceInstance",
			Handler:       _WorkspacesService_ListenToWorkspaceInstance_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "ListenToImageBuildLogs",
			Handler:       _WorkspacesService_ListenToImageBuildLogs_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "StopWorkspace",
			Handler:       _WorkspacesService_StopWorkspace_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "gitpod/v1/workspaces.proto",
}
