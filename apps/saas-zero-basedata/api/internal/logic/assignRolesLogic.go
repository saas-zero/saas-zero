package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type AssignRolesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAssignRolesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignRolesLogic {
	return &AssignRolesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AssignRolesLogic) AssignRoles(req *types.UserReq) (*types.BaseResp, error) {
	resp, err := l.svcCtx.SysUsers.AssignRoles(l.ctx, &apps.UserReq{
		Id:      proto.Int64(parseId(req.Id)),
		RoleIds: parseIds(req.RoleIds),
	})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg}, nil
}
