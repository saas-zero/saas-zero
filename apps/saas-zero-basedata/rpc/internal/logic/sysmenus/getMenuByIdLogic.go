package sysmenuslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent/sysmenu"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetMenuByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMenuByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMenuByIdLogic {
	return &GetMenuByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMenuByIdLogic) GetMenuById(in *apps.IdReq) (*apps.MenuResp, error) {
	m, err := l.svcCtx.DB.SysMenu.ActiveQuery().
		Where(sysmenu.IDEQ(in.GetId())).
		Only(l.ctx)
	if err != nil {
		return nil, err
	}
	return &apps.MenuResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: menuToResp(m),
	}, nil
}
