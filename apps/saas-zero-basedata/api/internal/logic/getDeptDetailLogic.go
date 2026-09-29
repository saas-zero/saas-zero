package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeptDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDeptDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeptDetailLogic {
	return &GetDeptDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDeptDetailLogic) GetDeptDetail(req *types.IdReq) (*types.BaseResp, error) {
	resp, err := l.svcCtx.SysDepts.GetDeptById(l.ctx, &apps.IdReq{Id: parseId(req.Id)})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: toDeptInfo(resp.GetData())}, nil
}
