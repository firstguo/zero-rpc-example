package logic

import (
	"context"

	user_auth "zero-rpc-example/buf_proto_example/gen/go/tripo/user_auth/v1"
	"zero-rpc-example/services/user-auth/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPasswordLogic {
	return &ResetPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ResetPasswordLogic) ResetPassword(in *user_auth.ResetPasswordRequest) (*user_auth.ResetPasswordResponse, error) {
	// todo: add your logic here and delete this line

	return &user_auth.ResetPasswordResponse{}, nil
}
