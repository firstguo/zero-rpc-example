package logic

import (
	"context"

	user_auth "zero-rpc-example/buf_proto_example/gen/go/tripo/user_auth/v1"
	"zero-rpc-example/services/user-auth/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RefreshTokenLogic) RefreshToken(in *user_auth.RefreshTokenRequest) (*user_auth.RefreshTokenResponse, error) {
	// todo: add your logic here and delete this line

	return &user_auth.RefreshTokenResponse{}, nil
}
