package sysmenuslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysmenu"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/pagination"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMenuListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuListLogic {
	return &GetMenuListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMenuListLogic) GetMenuList(in *apps.MenuPageReq) (*apps.MenuListResp, error) {
	query := l.svcCtx.DB.SysMenu.ActiveQuery()
	if in.GetName() != "" {
		query = query.Where(sysmenu.NameContains(in.GetName()))
	}
	if in.GetStatus() != "" {
		query = query.Where(sysmenu.StatusEQ(sysmenu.Status(in.GetStatus())))
	}

	total, err := query.Count(l.ctx)
	if err != nil {
		return nil, err
	}

	_, size, offset := pagination.Normalize(int(in.GetPage()), int(in.GetSize()))

	menus, err := query.
		Offset(offset).
		Limit(size).
		Order(ent.Asc(sysmenu.FieldSort)).
		All(l.ctx)
	if err != nil {
		return nil, err
	}

	list := make([]*apps.Menu, len(menus))
	for i, m := range menus {
		list[i] = menuToResp(m)
	}
	return &apps.MenuListResp{
		Code:  int32(errno.Success.Code),
		Msg:   errno.Success.Msg,
		List:  list,
		Total: int64(total),
	}, nil
}
