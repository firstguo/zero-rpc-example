package main

import (
	"flag"
	"fmt"

	user "zero-rpc-example/buf_proto_example/gen/go/example/base/svr/user/v1"
	"zero-rpc-example/internal/common"
	"zero-rpc-example/internal/config"
	"zero-rpc-example/internal/server"
	"zero-rpc-example/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/example.base.svr.user.user.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	// Register JSON codec for gRPC SubContentType support
	common.RegisterJSONCodec()

	// Apply environment prefix and tag routing
	common.ApplyEnvAndTagRouting(&c.RpcServerConf, c.Meta)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		user.RegisterUserServer(grpcServer, server.NewUserServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
