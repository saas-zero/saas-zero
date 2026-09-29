package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type ResetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPasswordLogic {
	return &ResetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetPasswordLogic) ResetPassword(req *types.UserReq) (*types.BaseResp, error) {
	resp, err := l.svcCtx.SysUsers.ResetPassword(l.ctx, &apps.UserReq{
		Id:       proto.Int64(parseId(req.Id)),
		Password: proto.String(req.Password),
	})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg}, nil
}
