package logic

import (
	"context"

	example "zero-rpc-example/buf_proto_example/gen/go/example/v1"
	"zero-rpc-example/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserLogic) GetUser(in *example.GetUserRequest) (*example.GetUserResponse, error) {
	// todo: add your logic here and delete this line

	return &example.GetUserResponse{}, nil
}
