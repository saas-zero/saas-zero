package sysdictdataslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdict"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdictdata"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDictDataByDictKeyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDictDataByDictKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDictDataByDictKeyLogic {
	return &GetDictDataByDictKeyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDictDataByDictKeyLogic) GetDictDataByDictKey(in *apps.DictReq) (*apps.DictDataListResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)

	// First try tenant-specific dict
	dict, err := l.svcCtx.DB.SysDict.Query().
		Where(sysdict.KeyEQ(in.GetKey()), sysdict.TenantIDEQ(tenantId), sysdict.DeletedAtIsNil()).
		Only(l.ctx)
	if err != nil {
		// Fall back to system default dict
		dict, err = l.svcCtx.DB.SysDict.Query().
			Where(sysdict.KeyEQ(in.GetKey()), sysdict.TenantIDEQ(0), sysdict.DeletedAtIsNil()).
			Only(l.ctx)
		if err != nil {
			return nil, err
		}
	}

	items, err := l.svcCtx.DB.SysDictData.Query().
		Where(sysdictdata.DictIDEQ(dict.ID), sysdictdata.DeletedAtIsNil()).
		Order(ent.Asc(sysdictdata.FieldCreatedAt)).
		All(l.ctx)
	if err != nil {
		return nil, err
	}

	list := make([]*apps.DictData, len(items))
	for i, d := range items {
		list[i] = dictDataToResp(d)
	}
	return &apps.DictDataListResp{
		Code:  int32(errno.Success.Code),
		Msg:   errno.Success.Msg,
		List:  list,
		Total: int64(len(list)),
	}, nil
}
