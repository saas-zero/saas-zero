package systenantslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/systenant"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/pagination"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type GetTenantListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantListLogic {
	return &GetTenantListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetTenantListLogic) GetTenantList(in *apps.TenantPageReq) (*apps.TenantListResp, error) {
	query := l.svcCtx.DB.SysTenant.ActiveQuery()
	if in.GetName() != "" {
		query = query.Where(systenant.NameContains(in.GetName()))
	}
	if in.GetCode() != "" {
		query = query.Where(systenant.CodeContains(in.GetCode()))
	}
	if in.GetStatus() != "" {
		query = query.Where(systenant.StatusEQ(systenant.Status(in.GetStatus())))
	}

	total, err := query.Count(l.ctx)
	if err != nil {
		return nil, err
	}

	_, size, offset := pagination.Normalize(int(in.GetPage()), int(in.GetSize()))

	tenants, err := query.
		Offset(offset).
		Limit(size).
		Order(ent.Asc(systenant.FieldCreatedAt)).
		WithSysPackage().
		All(l.ctx)
	if err != nil {
		return nil, err
	}

	// 批量加载管理员用户名（admin_id → username）
	adminNames := make(map[int64]string, len(tenants))
	adminIDs := make([]int64, 0, len(tenants))
	for _, t := range tenants {
		if t.AdminID > 0 {
			adminIDs = append(adminIDs, t.AdminID)
		}
	}
	if len(adminIDs) > 0 {
		admins, err := l.svcCtx.DB.SysUser.Query().Where(sysuser.IDIn(adminIDs...)).All(l.ctx)
		if err == nil {
			for _, u := range admins {
				adminNames[u.ID] = u.Username
			}
		}
	}

	list := make([]*apps.Tenant, len(tenants))
	for i, t := range tenants {
		resp := tenantToResp(t)
		if n, ok := adminNames[t.AdminID]; ok {
			resp.AdminName = proto.String(n)
		}
		list[i] = resp
	}
	return &apps.TenantListResp{
		Code:  int32(errno.Success.Code),
		Msg:   errno.Success.Msg,
		List:  list,
		Total: int64(total),
	}, nil
}
