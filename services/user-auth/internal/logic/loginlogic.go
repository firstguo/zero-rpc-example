package logic

import (
	"context"

	user_auth "zero-rpc-example/buf_proto_example/gen/go/tripo/user_auth/v1"
	"zero-rpc-example/services/user-auth/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *user_auth.LoginRequest) (*user_auth.LoginResponse, error) {
	// todo: add your logic here and delete this line

	return &user_auth.LoginResponse{}, nil
}
