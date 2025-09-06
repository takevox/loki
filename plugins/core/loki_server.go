package main

import (
	"context"

	"connectrpc.com/connect"
	lokiv1 "github.com/takevox/loki/gen/loki/v1"
)

type pluginServer struct{}

func (s *pluginServer) Initialize(ctx context.Context, req *connect.Request[lokiv1.InitializeRequest]) (*connect.Response[lokiv1.InitializeResponse], error) {
	return nil, nil
}
