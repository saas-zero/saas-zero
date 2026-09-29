package sysmenuslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent/sysmenu"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateMenuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMenuLogic {
	return &UpdateMenuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateMenuLogic) UpdateMenu(in *apps.MenuReq) (*apps.MenuResp, error) {
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentUserId(l.ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	update := l.svcCtx.DB.SysMenu.UpdateOneID(in.GetId())
	if in.MenuType != nil {
		update.SetMenuType(sysmenu.MenuType(in.GetMenuType()))
	}
	if in.Name != nil {
		update.SetName(in.GetName())
	}
	if in.ParentId != nil && in.GetParentId() > 0 {
		update.SetParentID(in.GetParentId())
	}
	if in.Component != nil {
		update.SetComponent(in.GetComponent())
	}
	if in.Path != nil {
		update.SetPath(in.GetPath())
	}
	if in.Icon != nil {
		update.SetIcon(in.GetIcon())
	}
	if in.IsRedirect != nil {
		update.SetIsRedirect(in.GetIsRedirect())
	}
	if in.Redirect != nil {
		update.SetRedirect(in.GetRedirect())
	}
	if in.Hidden != nil {
		update.SetHidden(in.GetHidden())
	}
	if in.Status != nil {
		update.SetStatus(sysmenu.Status(in.GetStatus()))
	}
	if in.Sort != nil {
		update.SetSort(uint32(in.GetSort()))
	}
	if in.Remark != nil {
		update.SetRemark(in.GetRemark())
	}

	result, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	m, err := l.svcCtx.DB.SysMenu.Query().Where(sysmenu.IDEQ(result.ID)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.MenuResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: menuToResp(m),
	}, nil
}
