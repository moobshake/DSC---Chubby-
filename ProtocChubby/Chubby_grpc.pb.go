// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package protocchubby

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

// NodeCommControlServiceClient is the client API for NodeCommControlService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NodeCommControlServiceClient interface {
	SendControlMessage(ctx context.Context, in *ControlMessage, opts ...grpc.CallOption) (*ControlMessage, error)
	Shutdown(ctx context.Context, in *ControlMessage, opts ...grpc.CallOption) (*ControlMessage, error)
}

type nodeCommControlServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNodeCommControlServiceClient(cc grpc.ClientConnInterface) NodeCommControlServiceClient {
	return &nodeCommControlServiceClient{cc}
}

func (c *nodeCommControlServiceClient) SendControlMessage(ctx context.Context, in *ControlMessage, opts ...grpc.CallOption) (*ControlMessage, error) {
	out := new(ControlMessage)
	err := c.cc.Invoke(ctx, "/main.NodeCommControlService/SendControlMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeCommControlServiceClient) Shutdown(ctx context.Context, in *ControlMessage, opts ...grpc.CallOption) (*ControlMessage, error) {
	out := new(ControlMessage)
	err := c.cc.Invoke(ctx, "/main.NodeCommControlService/Shutdown", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NodeCommControlServiceServer is the server API for NodeCommControlService service.
// All implementations must embed UnimplementedNodeCommControlServiceServer
// for forward compatibility
type NodeCommControlServiceServer interface {
	SendControlMessage(context.Context, *ControlMessage) (*ControlMessage, error)
	Shutdown(context.Context, *ControlMessage) (*ControlMessage, error)
	mustEmbedUnimplementedNodeCommControlServiceServer()
}

// UnimplementedNodeCommControlServiceServer must be embedded to have forward compatible implementations.
type UnimplementedNodeCommControlServiceServer struct {
}

func (UnimplementedNodeCommControlServiceServer) SendControlMessage(context.Context, *ControlMessage) (*ControlMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendControlMessage not implemented")
}
func (UnimplementedNodeCommControlServiceServer) Shutdown(context.Context, *ControlMessage) (*ControlMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Shutdown not implemented")
}
func (UnimplementedNodeCommControlServiceServer) mustEmbedUnimplementedNodeCommControlServiceServer() {
}

// UnsafeNodeCommControlServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NodeCommControlServiceServer will
// result in compilation errors.
type UnsafeNodeCommControlServiceServer interface {
	mustEmbedUnimplementedNodeCommControlServiceServer()
}

func RegisterNodeCommControlServiceServer(s grpc.ServiceRegistrar, srv NodeCommControlServiceServer) {
	s.RegisterService(&NodeCommControlService_ServiceDesc, srv)
}

func _NodeCommControlService_SendControlMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ControlMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeCommControlServiceServer).SendControlMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.NodeCommControlService/SendControlMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeCommControlServiceServer).SendControlMessage(ctx, req.(*ControlMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeCommControlService_Shutdown_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ControlMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeCommControlServiceServer).Shutdown(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.NodeCommControlService/Shutdown",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeCommControlServiceServer).Shutdown(ctx, req.(*ControlMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// NodeCommControlService_ServiceDesc is the grpc.ServiceDesc for NodeCommControlService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var NodeCommControlService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "main.NodeCommControlService",
	HandlerType: (*NodeCommControlServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendControlMessage",
			Handler:    _NodeCommControlService_SendControlMessage_Handler,
		},
		{
			MethodName: "Shutdown",
			Handler:    _NodeCommControlService_Shutdown_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "ProtocChubby/Chubby.proto",
}

// NodeCommPeerServiceClient is the client API for NodeCommPeerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NodeCommPeerServiceClient interface {
	KeepAlive(ctx context.Context, in *NodeMessage, opts ...grpc.CallOption) (*NodeMessage, error)
	SendMessage(ctx context.Context, in *NodeMessage, opts ...grpc.CallOption) (*NodeMessage, error)
	SendCoordinationMessage(ctx context.Context, in *CoordinationMessage, opts ...grpc.CallOption) (*CoordinationMessage, error)
}

type nodeCommPeerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNodeCommPeerServiceClient(cc grpc.ClientConnInterface) NodeCommPeerServiceClient {
	return &nodeCommPeerServiceClient{cc}
}

func (c *nodeCommPeerServiceClient) KeepAlive(ctx context.Context, in *NodeMessage, opts ...grpc.CallOption) (*NodeMessage, error) {
	out := new(NodeMessage)
	err := c.cc.Invoke(ctx, "/main.NodeCommPeerService/KeepAlive", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeCommPeerServiceClient) SendMessage(ctx context.Context, in *NodeMessage, opts ...grpc.CallOption) (*NodeMessage, error) {
	out := new(NodeMessage)
	err := c.cc.Invoke(ctx, "/main.NodeCommPeerService/SendMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeCommPeerServiceClient) SendCoordinationMessage(ctx context.Context, in *CoordinationMessage, opts ...grpc.CallOption) (*CoordinationMessage, error) {
	out := new(CoordinationMessage)
	err := c.cc.Invoke(ctx, "/main.NodeCommPeerService/SendCoordinationMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NodeCommPeerServiceServer is the server API for NodeCommPeerService service.
// All implementations must embed UnimplementedNodeCommPeerServiceServer
// for forward compatibility
type NodeCommPeerServiceServer interface {
	KeepAlive(context.Context, *NodeMessage) (*NodeMessage, error)
	SendMessage(context.Context, *NodeMessage) (*NodeMessage, error)
	SendCoordinationMessage(context.Context, *CoordinationMessage) (*CoordinationMessage, error)
	mustEmbedUnimplementedNodeCommPeerServiceServer()
}

// UnimplementedNodeCommPeerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedNodeCommPeerServiceServer struct {
}

func (UnimplementedNodeCommPeerServiceServer) KeepAlive(context.Context, *NodeMessage) (*NodeMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method KeepAlive not implemented")
}
func (UnimplementedNodeCommPeerServiceServer) SendMessage(context.Context, *NodeMessage) (*NodeMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMessage not implemented")
}
func (UnimplementedNodeCommPeerServiceServer) SendCoordinationMessage(context.Context, *CoordinationMessage) (*CoordinationMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendCoordinationMessage not implemented")
}
func (UnimplementedNodeCommPeerServiceServer) mustEmbedUnimplementedNodeCommPeerServiceServer() {}

// UnsafeNodeCommPeerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NodeCommPeerServiceServer will
// result in compilation errors.
type UnsafeNodeCommPeerServiceServer interface {
	mustEmbedUnimplementedNodeCommPeerServiceServer()
}

func RegisterNodeCommPeerServiceServer(s grpc.ServiceRegistrar, srv NodeCommPeerServiceServer) {
	s.RegisterService(&NodeCommPeerService_ServiceDesc, srv)
}

func _NodeCommPeerService_KeepAlive_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeCommPeerServiceServer).KeepAlive(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.NodeCommPeerService/KeepAlive",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeCommPeerServiceServer).KeepAlive(ctx, req.(*NodeMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeCommPeerService_SendMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeCommPeerServiceServer).SendMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.NodeCommPeerService/SendMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeCommPeerServiceServer).SendMessage(ctx, req.(*NodeMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeCommPeerService_SendCoordinationMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CoordinationMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeCommPeerServiceServer).SendCoordinationMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.NodeCommPeerService/SendCoordinationMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeCommPeerServiceServer).SendCoordinationMessage(ctx, req.(*CoordinationMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// NodeCommPeerService_ServiceDesc is the grpc.ServiceDesc for NodeCommPeerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var NodeCommPeerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "main.NodeCommPeerService",
	HandlerType: (*NodeCommPeerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "KeepAlive",
			Handler:    _NodeCommPeerService_KeepAlive_Handler,
		},
		{
			MethodName: "SendMessage",
			Handler:    _NodeCommPeerService_SendMessage_Handler,
		},
		{
			MethodName: "SendCoordinationMessage",
			Handler:    _NodeCommPeerService_SendCoordinationMessage_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "ProtocChubby/Chubby.proto",
}

// NodeCommListeningServiceClient is the client API for NodeCommListeningService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NodeCommListeningServiceClient interface {
	SendClientMessage(ctx context.Context, in *ClientMessage, opts ...grpc.CallOption) (*ClientMessage, error)
	SendReadRequest(ctx context.Context, in *ClientMessage, opts ...grpc.CallOption) (NodeCommListeningService_SendReadRequestClient, error)
	SendWriteRequest(ctx context.Context, opts ...grpc.CallOption) (NodeCommListeningService_SendWriteRequestClient, error)
}

type nodeCommListeningServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNodeCommListeningServiceClient(cc grpc.ClientConnInterface) NodeCommListeningServiceClient {
	return &nodeCommListeningServiceClient{cc}
}

func (c *nodeCommListeningServiceClient) SendClientMessage(ctx context.Context, in *ClientMessage, opts ...grpc.CallOption) (*ClientMessage, error) {
	out := new(ClientMessage)
	err := c.cc.Invoke(ctx, "/main.NodeCommListeningService/SendClientMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeCommListeningServiceClient) SendReadRequest(ctx context.Context, in *ClientMessage, opts ...grpc.CallOption) (NodeCommListeningService_SendReadRequestClient, error) {
	stream, err := c.cc.NewStream(ctx, &NodeCommListeningService_ServiceDesc.Streams[0], "/main.NodeCommListeningService/SendReadRequest", opts...)
	if err != nil {
		return nil, err
	}
	x := &nodeCommListeningServiceSendReadRequestClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type NodeCommListeningService_SendReadRequestClient interface {
	Recv() (*ClientMessage, error)
	grpc.ClientStream
}

type nodeCommListeningServiceSendReadRequestClient struct {
	grpc.ClientStream
}

func (x *nodeCommListeningServiceSendReadRequestClient) Recv() (*ClientMessage, error) {
	m := new(ClientMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *nodeCommListeningServiceClient) SendWriteRequest(ctx context.Context, opts ...grpc.CallOption) (NodeCommListeningService_SendWriteRequestClient, error) {
	stream, err := c.cc.NewStream(ctx, &NodeCommListeningService_ServiceDesc.Streams[1], "/main.NodeCommListeningService/SendWriteRequest", opts...)
	if err != nil {
		return nil, err
	}
	x := &nodeCommListeningServiceSendWriteRequestClient{stream}
	return x, nil
}

type NodeCommListeningService_SendWriteRequestClient interface {
	Send(*ClientMessage) error
	CloseAndRecv() (*ClientMessage, error)
	grpc.ClientStream
}

type nodeCommListeningServiceSendWriteRequestClient struct {
	grpc.ClientStream
}

func (x *nodeCommListeningServiceSendWriteRequestClient) Send(m *ClientMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *nodeCommListeningServiceSendWriteRequestClient) CloseAndRecv() (*ClientMessage, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(ClientMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// NodeCommListeningServiceServer is the server API for NodeCommListeningService service.
// All implementations must embed UnimplementedNodeCommListeningServiceServer
// for forward compatibility
type NodeCommListeningServiceServer interface {
	SendClientMessage(context.Context, *ClientMessage) (*ClientMessage, error)
	SendReadRequest(*ClientMessage, NodeCommListeningService_SendReadRequestServer) error
	SendWriteRequest(NodeCommListeningService_SendWriteRequestServer) error
	mustEmbedUnimplementedNodeCommListeningServiceServer()
}

// UnimplementedNodeCommListeningServiceServer must be embedded to have forward compatible implementations.
type UnimplementedNodeCommListeningServiceServer struct {
}

func (UnimplementedNodeCommListeningServiceServer) SendClientMessage(context.Context, *ClientMessage) (*ClientMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendClientMessage not implemented")
}
func (UnimplementedNodeCommListeningServiceServer) SendReadRequest(*ClientMessage, NodeCommListeningService_SendReadRequestServer) error {
	return status.Errorf(codes.Unimplemented, "method SendReadRequest not implemented")
}
func (UnimplementedNodeCommListeningServiceServer) SendWriteRequest(NodeCommListeningService_SendWriteRequestServer) error {
	return status.Errorf(codes.Unimplemented, "method SendWriteRequest not implemented")
}
func (UnimplementedNodeCommListeningServiceServer) mustEmbedUnimplementedNodeCommListeningServiceServer() {
}

// UnsafeNodeCommListeningServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NodeCommListeningServiceServer will
// result in compilation errors.
type UnsafeNodeCommListeningServiceServer interface {
	mustEmbedUnimplementedNodeCommListeningServiceServer()
}

func RegisterNodeCommListeningServiceServer(s grpc.ServiceRegistrar, srv NodeCommListeningServiceServer) {
	s.RegisterService(&NodeCommListeningService_ServiceDesc, srv)
}

func _NodeCommListeningService_SendClientMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClientMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeCommListeningServiceServer).SendClientMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.NodeCommListeningService/SendClientMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeCommListeningServiceServer).SendClientMessage(ctx, req.(*ClientMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeCommListeningService_SendReadRequest_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ClientMessage)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(NodeCommListeningServiceServer).SendReadRequest(m, &nodeCommListeningServiceSendReadRequestServer{stream})
}

type NodeCommListeningService_SendReadRequestServer interface {
	Send(*ClientMessage) error
	grpc.ServerStream
}

type nodeCommListeningServiceSendReadRequestServer struct {
	grpc.ServerStream
}

func (x *nodeCommListeningServiceSendReadRequestServer) Send(m *ClientMessage) error {
	return x.ServerStream.SendMsg(m)
}

func _NodeCommListeningService_SendWriteRequest_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(NodeCommListeningServiceServer).SendWriteRequest(&nodeCommListeningServiceSendWriteRequestServer{stream})
}

type NodeCommListeningService_SendWriteRequestServer interface {
	SendAndClose(*ClientMessage) error
	Recv() (*ClientMessage, error)
	grpc.ServerStream
}

type nodeCommListeningServiceSendWriteRequestServer struct {
	grpc.ServerStream
}

func (x *nodeCommListeningServiceSendWriteRequestServer) SendAndClose(m *ClientMessage) error {
	return x.ServerStream.SendMsg(m)
}

func (x *nodeCommListeningServiceSendWriteRequestServer) Recv() (*ClientMessage, error) {
	m := new(ClientMessage)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// NodeCommListeningService_ServiceDesc is the grpc.ServiceDesc for NodeCommListeningService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var NodeCommListeningService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "main.NodeCommListeningService",
	HandlerType: (*NodeCommListeningServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendClientMessage",
			Handler:    _NodeCommListeningService_SendClientMessage_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SendReadRequest",
			Handler:       _NodeCommListeningService_SendReadRequest_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "SendWriteRequest",
			Handler:       _NodeCommListeningService_SendWriteRequest_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "ProtocChubby/Chubby.proto",
}

// ClientControlServiceClient is the client API for ClientControlService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ClientControlServiceClient interface {
	SendControlMessage(ctx context.Context, in *ControlMessage, opts ...grpc.CallOption) (*ControlMessage, error)
	Shutdown(ctx context.Context, in *ControlMessage, opts ...grpc.CallOption) (*ControlMessage, error)
}

type clientControlServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewClientControlServiceClient(cc grpc.ClientConnInterface) ClientControlServiceClient {
	return &clientControlServiceClient{cc}
}

func (c *clientControlServiceClient) SendControlMessage(ctx context.Context, in *ControlMessage, opts ...grpc.CallOption) (*ControlMessage, error) {
	out := new(ControlMessage)
	err := c.cc.Invoke(ctx, "/main.ClientControlService/SendControlMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientControlServiceClient) Shutdown(ctx context.Context, in *ControlMessage, opts ...grpc.CallOption) (*ControlMessage, error) {
	out := new(ControlMessage)
	err := c.cc.Invoke(ctx, "/main.ClientControlService/Shutdown", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ClientControlServiceServer is the server API for ClientControlService service.
// All implementations must embed UnimplementedClientControlServiceServer
// for forward compatibility
type ClientControlServiceServer interface {
	SendControlMessage(context.Context, *ControlMessage) (*ControlMessage, error)
	Shutdown(context.Context, *ControlMessage) (*ControlMessage, error)
	mustEmbedUnimplementedClientControlServiceServer()
}

// UnimplementedClientControlServiceServer must be embedded to have forward compatible implementations.
type UnimplementedClientControlServiceServer struct {
}

func (UnimplementedClientControlServiceServer) SendControlMessage(context.Context, *ControlMessage) (*ControlMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendControlMessage not implemented")
}
func (UnimplementedClientControlServiceServer) Shutdown(context.Context, *ControlMessage) (*ControlMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Shutdown not implemented")
}
func (UnimplementedClientControlServiceServer) mustEmbedUnimplementedClientControlServiceServer() {}

// UnsafeClientControlServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ClientControlServiceServer will
// result in compilation errors.
type UnsafeClientControlServiceServer interface {
	mustEmbedUnimplementedClientControlServiceServer()
}

func RegisterClientControlServiceServer(s grpc.ServiceRegistrar, srv ClientControlServiceServer) {
	s.RegisterService(&ClientControlService_ServiceDesc, srv)
}

func _ClientControlService_SendControlMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ControlMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientControlServiceServer).SendControlMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.ClientControlService/SendControlMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientControlServiceServer).SendControlMessage(ctx, req.(*ControlMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientControlService_Shutdown_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ControlMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientControlServiceServer).Shutdown(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.ClientControlService/Shutdown",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientControlServiceServer).Shutdown(ctx, req.(*ControlMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// ClientControlService_ServiceDesc is the grpc.ServiceDesc for ClientControlService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ClientControlService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "main.ClientControlService",
	HandlerType: (*ClientControlServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendControlMessage",
			Handler:    _ClientControlService_SendControlMessage_Handler,
		},
		{
			MethodName: "Shutdown",
			Handler:    _ClientControlService_Shutdown_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "ProtocChubby/Chubby.proto",
}

// ClientListeningServiceClient is the client API for ClientListeningService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ClientListeningServiceClient interface {
	SendClientMessage(ctx context.Context, in *ClientMessage, opts ...grpc.CallOption) (*ClientMessage, error)
	SendEventMessage(ctx context.Context, in *EventMessage, opts ...grpc.CallOption) (*EventMessage, error)
}

type clientListeningServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewClientListeningServiceClient(cc grpc.ClientConnInterface) ClientListeningServiceClient {
	return &clientListeningServiceClient{cc}
}

func (c *clientListeningServiceClient) SendClientMessage(ctx context.Context, in *ClientMessage, opts ...grpc.CallOption) (*ClientMessage, error) {
	out := new(ClientMessage)
	err := c.cc.Invoke(ctx, "/main.ClientListeningService/SendClientMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientListeningServiceClient) SendEventMessage(ctx context.Context, in *EventMessage, opts ...grpc.CallOption) (*EventMessage, error) {
	out := new(EventMessage)
	err := c.cc.Invoke(ctx, "/main.ClientListeningService/SendEventMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ClientListeningServiceServer is the server API for ClientListeningService service.
// All implementations must embed UnimplementedClientListeningServiceServer
// for forward compatibility
type ClientListeningServiceServer interface {
	SendClientMessage(context.Context, *ClientMessage) (*ClientMessage, error)
	SendEventMessage(context.Context, *EventMessage) (*EventMessage, error)
	mustEmbedUnimplementedClientListeningServiceServer()
}

// UnimplementedClientListeningServiceServer must be embedded to have forward compatible implementations.
type UnimplementedClientListeningServiceServer struct {
}

func (UnimplementedClientListeningServiceServer) SendClientMessage(context.Context, *ClientMessage) (*ClientMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendClientMessage not implemented")
}
func (UnimplementedClientListeningServiceServer) SendEventMessage(context.Context, *EventMessage) (*EventMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendEventMessage not implemented")
}
func (UnimplementedClientListeningServiceServer) mustEmbedUnimplementedClientListeningServiceServer() {
}

// UnsafeClientListeningServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ClientListeningServiceServer will
// result in compilation errors.
type UnsafeClientListeningServiceServer interface {
	mustEmbedUnimplementedClientListeningServiceServer()
}

func RegisterClientListeningServiceServer(s grpc.ServiceRegistrar, srv ClientListeningServiceServer) {
	s.RegisterService(&ClientListeningService_ServiceDesc, srv)
}

func _ClientListeningService_SendClientMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ClientMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientListeningServiceServer).SendClientMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.ClientListeningService/SendClientMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientListeningServiceServer).SendClientMessage(ctx, req.(*ClientMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientListeningService_SendEventMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EventMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientListeningServiceServer).SendEventMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/main.ClientListeningService/SendEventMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientListeningServiceServer).SendEventMessage(ctx, req.(*EventMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// ClientListeningService_ServiceDesc is the grpc.ServiceDesc for ClientListeningService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ClientListeningService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "main.ClientListeningService",
	HandlerType: (*ClientListeningServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendClientMessage",
			Handler:    _ClientListeningService_SendClientMessage_Handler,
		},
		{
			MethodName: "SendEventMessage",
			Handler:    _ClientListeningService_SendEventMessage_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "ProtocChubby/Chubby.proto",
}
