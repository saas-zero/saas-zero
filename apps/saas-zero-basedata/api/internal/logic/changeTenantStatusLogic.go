package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type ChangeTenantStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChangeTenantStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeTenantStatusLogic {
	return &ChangeTenantStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangeTenantStatusLogic) ChangeTenantStatus(req *types.TenantReq) (*types.BaseResp, error) {
	resp, err := l.svcCtx.SysTenants.ChangeTenantStatus(l.ctx, &apps.TenantReq{
		Id:     proto.Int64(parseId(req.Id)),
		Status: proto.String(req.Status),
	})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg}, nil
}
