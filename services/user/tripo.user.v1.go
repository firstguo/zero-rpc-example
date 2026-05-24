package main

import (
	"flag"
	"fmt"

	user "zero-rpc-example/buf_proto_example/gen/go/tripo/user/v1"
	"zero-rpc-example/internal/common"
	"zero-rpc-example/services/user/internal/config"
	"zero-rpc-example/services/user/internal/server"
	"zero-rpc-example/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/tripo.user.v1.yaml", "the config file")

func main() {
	flag.Parse()

	common.RegisterJSONCodec()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	common.ApplyEnvAndTagRouting(&c.RpcServerConf, c.Meta)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		user.RegisterUserServiceServer(grpcServer, server.NewUserServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
