package logic

import (
	"context"

	user_auth "zero-rpc-example/buf_proto_example/gen/go/tripo/user_auth/v1"
	"zero-rpc-example/services/user-auth/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifyTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyTokenLogic {
	return &VerifyTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *VerifyTokenLogic) VerifyToken(in *user_auth.VerifyTokenRequest) (*user_auth.VerifyTokenResponse, error) {
	// todo: add your logic here and delete this line

	return &user_auth.VerifyTokenResponse{}, nil
}
