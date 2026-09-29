package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeptTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDeptTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeptTreeLogic {
	return &GetDeptTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDeptTreeLogic) GetDeptTree() (*types.BaseResp, error) {
	resp, err := l.svcCtx.SysDepts.GetDeptTree(l.ctx, &apps.EmptyReq{})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: toDeptInfoList(resp.GetData())}, nil
}
