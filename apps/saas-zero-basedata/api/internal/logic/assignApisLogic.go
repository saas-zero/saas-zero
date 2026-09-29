package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type AssignApisLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignApisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignApisLogic {
	return &AssignApisLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AssignApisLogic) AssignApis(req *types.RoleReq) (*types.BaseResp, error) {
	resp, err := l.svcCtx.SysRoles.AssignApis(l.ctx, &apps.RoleReq{
		Id:     proto.Int64(parseId(req.Id)),
		ApiIds: parseIds(req.ApiIds),
	})
	if err != nil {
		return nil, err
	}
	if l.svcCtx.Enforcer != nil {
		if err := l.svcCtx.Enforcer.LoadPolicy(); err != nil {
			logx.Errorf("assignApis: failed to reload casbin policies: %v", err)
		}
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg}, nil
}
