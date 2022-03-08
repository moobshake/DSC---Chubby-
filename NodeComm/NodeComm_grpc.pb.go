// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package nodecomm

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

// NodeCommServiceClient is the client API for NodeCommService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NodeCommServiceClient interface {
	KeepAlive(ctx context.Context, in *NodeMessage, opts ...grpc.CallOption) (*NodeMessage, error)
	SendMessage(ctx context.Context, in *NodeMessage, opts ...grpc.CallOption) (*NodeMessage, error)
	SendCoordinationMessage(ctx context.Context, in *CoordinationMessage, opts ...grpc.CallOption) (*CoordinationMessage, error)
	Shutdown(ctx context.Context, in *ControlMessage, opts ...grpc.CallOption) (*ControlMessage, error)
	SendControlMessage(ctx context.Context, in *ControlMessage, opts ...grpc.CallOption) (*ControlMessage, error)
}

type nodeCommServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNodeCommServiceClient(cc grpc.ClientConnInterface) NodeCommServiceClient {
	return &nodeCommServiceClient{cc}
}

func (c *nodeCommServiceClient) KeepAlive(ctx context.Context, in *NodeMessage, opts ...grpc.CallOption) (*NodeMessage, error) {
	out := new(NodeMessage)
	err := c.cc.Invoke(ctx, "/NodeComm.NodeCommService/KeepAlive", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeCommServiceClient) SendMessage(ctx context.Context, in *NodeMessage, opts ...grpc.CallOption) (*NodeMessage, error) {
	out := new(NodeMessage)
	err := c.cc.Invoke(ctx, "/NodeComm.NodeCommService/SendMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeCommServiceClient) SendCoordinationMessage(ctx context.Context, in *CoordinationMessage, opts ...grpc.CallOption) (*CoordinationMessage, error) {
	out := new(CoordinationMessage)
	err := c.cc.Invoke(ctx, "/NodeComm.NodeCommService/SendCoordinationMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeCommServiceClient) Shutdown(ctx context.Context, in *ControlMessage, opts ...grpc.CallOption) (*ControlMessage, error) {
	out := new(ControlMessage)
	err := c.cc.Invoke(ctx, "/NodeComm.NodeCommService/Shutdown", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nodeCommServiceClient) SendControlMessage(ctx context.Context, in *ControlMessage, opts ...grpc.CallOption) (*ControlMessage, error) {
	out := new(ControlMessage)
	err := c.cc.Invoke(ctx, "/NodeComm.NodeCommService/SendControlMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NodeCommServiceServer is the server API for NodeCommService service.
// All implementations must embed UnimplementedNodeCommServiceServer
// for forward compatibility
type NodeCommServiceServer interface {
	KeepAlive(context.Context, *NodeMessage) (*NodeMessage, error)
	SendMessage(context.Context, *NodeMessage) (*NodeMessage, error)
	SendCoordinationMessage(context.Context, *CoordinationMessage) (*CoordinationMessage, error)
	Shutdown(context.Context, *ControlMessage) (*ControlMessage, error)
	SendControlMessage(context.Context, *ControlMessage) (*ControlMessage, error)
	mustEmbedUnimplementedNodeCommServiceServer()
}

// UnimplementedNodeCommServiceServer must be embedded to have forward compatible implementations.
type UnimplementedNodeCommServiceServer struct {
}

func (UnimplementedNodeCommServiceServer) KeepAlive(context.Context, *NodeMessage) (*NodeMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method KeepAlive not implemented")
}
func (UnimplementedNodeCommServiceServer) SendMessage(context.Context, *NodeMessage) (*NodeMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMessage not implemented")
}
func (UnimplementedNodeCommServiceServer) SendCoordinationMessage(context.Context, *CoordinationMessage) (*CoordinationMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendCoordinationMessage not implemented")
}
func (UnimplementedNodeCommServiceServer) Shutdown(context.Context, *ControlMessage) (*ControlMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Shutdown not implemented")
}
func (UnimplementedNodeCommServiceServer) SendControlMessage(context.Context, *ControlMessage) (*ControlMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendControlMessage not implemented")
}
func (UnimplementedNodeCommServiceServer) mustEmbedUnimplementedNodeCommServiceServer() {}

// UnsafeNodeCommServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NodeCommServiceServer will
// result in compilation errors.
type UnsafeNodeCommServiceServer interface {
	mustEmbedUnimplementedNodeCommServiceServer()
}

func RegisterNodeCommServiceServer(s grpc.ServiceRegistrar, srv NodeCommServiceServer) {
	s.RegisterService(&NodeCommService_ServiceDesc, srv)
}

func _NodeCommService_KeepAlive_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeCommServiceServer).KeepAlive(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/NodeComm.NodeCommService/KeepAlive",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeCommServiceServer).KeepAlive(ctx, req.(*NodeMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeCommService_SendMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NodeMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeCommServiceServer).SendMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/NodeComm.NodeCommService/SendMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeCommServiceServer).SendMessage(ctx, req.(*NodeMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeCommService_SendCoordinationMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CoordinationMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeCommServiceServer).SendCoordinationMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/NodeComm.NodeCommService/SendCoordinationMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeCommServiceServer).SendCoordinationMessage(ctx, req.(*CoordinationMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeCommService_Shutdown_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ControlMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeCommServiceServer).Shutdown(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/NodeComm.NodeCommService/Shutdown",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeCommServiceServer).Shutdown(ctx, req.(*ControlMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _NodeCommService_SendControlMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ControlMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NodeCommServiceServer).SendControlMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/NodeComm.NodeCommService/SendControlMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NodeCommServiceServer).SendControlMessage(ctx, req.(*ControlMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// NodeCommService_ServiceDesc is the grpc.ServiceDesc for NodeCommService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var NodeCommService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "NodeComm.NodeCommService",
	HandlerType: (*NodeCommServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "KeepAlive",
			Handler:    _NodeCommService_KeepAlive_Handler,
		},
		{
			MethodName: "SendMessage",
			Handler:    _NodeCommService_SendMessage_Handler,
		},
		{
			MethodName: "SendCoordinationMessage",
			Handler:    _NodeCommService_SendCoordinationMessage_Handler,
		},
		{
			MethodName: "Shutdown",
			Handler:    _NodeCommService_Shutdown_Handler,
		},
		{
			MethodName: "SendControlMessage",
			Handler:    _NodeCommService_SendControlMessage_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "NodeComm.proto",
}
