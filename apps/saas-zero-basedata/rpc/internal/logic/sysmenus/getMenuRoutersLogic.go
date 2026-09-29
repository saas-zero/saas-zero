package sysmenuslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysmenu"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuRoutersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMenuRoutersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuRoutersLogic {
	return &GetMenuRoutersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMenuRoutersLogic) GetMenuRouters(_ *apps.EmptyReq) (*apps.MenuTreeResp, error) {
	allMenus, err := l.svcCtx.DB.SysMenu.ActiveQuery().
		Order(ent.Asc(sysmenu.FieldSort)).
		All(l.ctx)
	if err != nil {
		return nil, err
	}

	tree := buildRouterTree(allMenus, 0)
	return &apps.MenuTreeResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: tree,
	}, nil
}
