package sysdictslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent/sysdict"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDictByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDictByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDictByIdLogic {
	return &GetDictByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDictByIdLogic) GetDictById(in *apps.IdReq) (*apps.DictResp, error) {
	d, err := l.svcCtx.DB.SysDict.Query().
		Where(sysdict.IDEQ(in.GetId()), sysdict.DeletedAtIsNil()).
		WithSysDictDatas().
		Only(l.ctx)
	if err != nil {
		return nil, err
	}
	return &apps.DictResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: dictToResp(d),
	}, nil
}
