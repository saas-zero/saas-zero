package sysapislogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent/sysapi"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetApiByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetApiByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetApiByIdLogic {
	return &GetApiByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetApiByIdLogic) GetApiById(in *apps.IdReq) (*apps.ApiResp, error) {
	a, err := l.svcCtx.DB.SysApi.ActiveQuery().
		Where(sysapi.IDEQ(in.GetId())).
		Only(l.ctx)
	if err != nil {
		return nil, err
	}
	return &apps.ApiResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: apiToResp(a),
	}, nil
}
