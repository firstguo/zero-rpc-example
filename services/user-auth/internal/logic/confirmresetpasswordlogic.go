package logic

import (
	"context"

	user_auth "zero-rpc-example/buf_proto_example/gen/go/tripo/user_auth/v1"
	"zero-rpc-example/services/user-auth/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConfirmResetPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewConfirmResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmResetPasswordLogic {
	return &ConfirmResetPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ConfirmResetPasswordLogic) ConfirmResetPassword(in *user_auth.ConfirmResetPasswordRequest) (*user_auth.ConfirmResetPasswordResponse, error) {
	// todo: add your logic here and delete this line

	return &user_auth.ConfirmResetPasswordResponse{}, nil
}
