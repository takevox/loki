package main

import (
	"context"

	"connectrpc.com/connect"
	lokiv1 "github.com/takevox/loki/gen/loki/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type PluginServer struct{}

func (s *PluginServer) Initialize(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *PluginServer) Terminate(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *PluginServer) GetNodeNameList(ctx context.Context, req *connect.Request[emptypb.Empty]) (*connect.Response[lokiv1.StringArray], error) {
	res := &lokiv1.StringArray{}
	return connect.NewResponse(res), nil
}

func (s *PluginServer) Stream(ctx context.Context, req *connect.Request[lokiv1.StreamPack]) (*connect.Response[lokiv1.StreamPack], error) {
	return nil, nil
}

func (s *PluginServer) Startup(ctx context.Context, req *connect.Request[lokiv1.HandlerRequest]) (*connect.Response[lokiv1.HandlerResponse], error) {
	return nil, nil
}

func (s *PluginServer) Shutdown(ctx context.Context, req *connect.Request[lokiv1.HandlerRequest]) (*connect.Response[lokiv1.HandlerResponse], error) {
	return nil, nil
}

func (s *PluginServer) PreProcess(ctx context.Context, req *connect.Request[lokiv1.HandlerRequest]) (*connect.Response[lokiv1.HandlerResponse], error) {
	return nil, nil
}

func (s *PluginServer) Process(ctx context.Context, req *connect.Request[lokiv1.HandlerRequest]) (*connect.Response[lokiv1.HandlerResponse], error) {
	return nil, nil
}

func (s *PluginServer) PostProcess(ctx context.Context, req *connect.Request[lokiv1.HandlerRequest]) (*connect.Response[lokiv1.HandlerResponse], error) {
	return nil, nil
}
