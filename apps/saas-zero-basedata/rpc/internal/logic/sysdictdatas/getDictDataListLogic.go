package sysdictdataslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdictdata"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/pagination"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDictDataListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDictDataListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDictDataListLogic {
	return &GetDictDataListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDictDataListLogic) GetDictDataList(in *apps.DictDataPageReq) (*apps.DictDataListResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)

	query := l.svcCtx.DB.SysDictData.TenantAwareQuery(tenantId)

	if in.GetDictId() > 0 {
		query = query.Where(sysdictdata.DictIDEQ(in.GetDictId()))
	}
	if in.GetKey() != "" {
		query = query.Where(sysdictdata.KeyContains(in.GetKey()))
	}
	if in.GetValue() != "" {
		query = query.Where(sysdictdata.ValueContains(in.GetValue()))
	}
	if in.GetStatus() != "" {
		query = query.Where(sysdictdata.StatusEQ(sysdictdata.Status(in.GetStatus())))
	}

	total, err := query.Count(l.ctx)
	if err != nil {
		return nil, err
	}

	_, size, offset := pagination.Normalize(int(in.GetPage()), int(in.GetSize()))

	items, err := query.
		Offset(offset).
		Limit(size).
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
		Total: int64(total),
	}, nil
}
