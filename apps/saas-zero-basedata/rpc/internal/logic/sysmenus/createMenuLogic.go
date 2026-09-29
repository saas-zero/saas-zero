package sysmenuslogic

import (
	"context"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/id"

	"github.com/saas-zero/saas-zero-basedata/ent/sysmenu"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type CreateMenuLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateMenuLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateMenuLogic {
	return &CreateMenuLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateMenuLogic) CreateMenu(in *apps.MenuReq) (*apps.MenuResp, error) {
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentUserId(l.ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	create := l.svcCtx.DB.SysMenu.Create().
		SetMenuType(sysmenu.MenuType(in.GetMenuType())).
		SetName(in.GetName()).
		SetComponent(in.GetComponent()).
		SetPath(in.GetPath()).
		SetIcon(in.GetIcon()).
		SetIsRedirect(in.GetIsRedirect()).
		SetRedirect(in.GetRedirect()).
		SetHidden(in.GetHidden()).
		SetStatus(sysmenu.Status(in.GetStatus())).
		SetSort(uint32(in.GetSort())).
		SetRemark(in.GetRemark())

	if in.GetParentId() > 0 {
		create.SetParentID(in.GetParentId())
	}

	result, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.MenuResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: &apps.Menu{
			Id:    proto.Int64(result.ID),
			IdStr: proto.String(id.ToString(result.ID)),
		},
	}, nil
}
