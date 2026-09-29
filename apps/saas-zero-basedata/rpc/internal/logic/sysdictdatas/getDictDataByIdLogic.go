package sysdictdataslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent/sysdictdata"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDictDataByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDictDataByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDictDataByIdLogic {
	return &GetDictDataByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDictDataByIdLogic) GetDictDataById(in *apps.IdReq) (*apps.DictDataResp, error) {
	d, err := l.svcCtx.DB.SysDictData.Query().
		Where(sysdictdata.IDEQ(in.GetId()), sysdictdata.DeletedAtIsNil()).
		Only(l.ctx)
	if err != nil {
		return nil, err
	}
	return &apps.DictDataResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: dictDataToResp(d),
	}, nil
}
