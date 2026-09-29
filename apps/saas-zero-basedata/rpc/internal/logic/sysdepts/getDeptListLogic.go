package sysdeptslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdept"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/pagination"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeptListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDeptListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeptListLogic {
	return &GetDeptListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDeptListLogic) GetDeptList(in *apps.DeptPageReq) (*apps.DeptListResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)

	query := l.svcCtx.DB.SysDept.TenantQuery(tenantId)
	if in.GetName() != "" {
		query = query.Where(sysdept.NameContains(in.GetName()))
	}
	if in.GetStatus() != "" {
		query = query.Where(sysdept.StatusEQ(sysdept.Status(in.GetStatus())))
	}

	total, err := query.Count(l.ctx)
	if err != nil {
		return nil, err
	}

	_, size, offset := pagination.Normalize(int(in.GetPage()), int(in.GetSize()))

	depts, err := query.
		Offset(offset).
		Limit(size).
		Order(ent.Asc(sysdept.FieldSort)).
		WithLeader().
		All(l.ctx)
	if err != nil {
		return nil, err
	}

	list := make([]*apps.Dept, len(depts))
	for i, d := range depts {
		list[i] = deptToResp(d)
	}
	return &apps.DeptListResp{
		Code:  int32(errno.Success.Code),
		Msg:   errno.Success.Msg,
		List:  list,
		Total: int64(total),
	}, nil
}
