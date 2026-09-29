package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type AssignMenusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignMenusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignMenusLogic {
	return &AssignMenusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AssignMenusLogic) AssignMenus(req *types.RoleReq) (*types.BaseResp, error) {
	resp, err := l.svcCtx.SysRoles.AssignMenus(l.ctx, &apps.RoleReq{
		Id:      proto.Int64(parseId(req.Id)),
		MenuIds: parseIds(req.MenuIds),
	})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg}, nil
}
