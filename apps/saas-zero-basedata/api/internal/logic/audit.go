package logic

import (
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	idutil "github.com/saas-zero/saas-zero-common/pkg/id"
	"github.com/saas-zero/saas-zero-common/pkg/timex"
)

// auditTime formats the millisecond timestamps used by the RPC layer for HTTP
// responses. Keeping this conversion at the API boundary gives every frontend
// page one stable, human-readable representation.
func auditTime(ts int64) string {
	return timex.FormatUnix(ts)
}

func toRoleInfo(r *apps.Role) *types.RoleInfo {
	if r == nil {
		return nil
	}
	return &types.RoleInfo{
		Id:          r.GetId(),
		IdStr:       r.GetIdStr(),
		Name:        r.GetName(),
		Code:        r.GetCode(),
		Status:      r.GetStatus(),
		Sort:        r.GetSort(),
		Remark:      r.GetRemark(),
		TenantId:    r.GetTenantId(),
		TenantIdStr: r.GetTenantIdStr(),
		MenuIds:     idutil.ToStrings(r.GetMenuIds()),
		ApiIds:      idutil.ToStrings(roleAPIIDs(r)),
		CreatedAt:   auditTime(r.GetCreatedAt()),
		CreatedBy:   r.GetCreatedBy(),
		UpdatedAt:   auditTime(r.GetUpdatedAt()),
		UpdatedBy:   r.GetUpdatedBy(),
	}
}

func roleAPIIDs(r *apps.Role) []int64 {
	return r.GetApiIds()
}

func toRoleInfoList(items []*apps.Role) []*types.RoleInfo {
	out := make([]*types.RoleInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toRoleInfo(item))
	}
	return out
}

func toDeptInfo(d *apps.Dept) *types.DeptInfo {
	if d == nil {
		return nil
	}
	children := make([]*types.DeptInfo, 0, len(d.GetChildren()))
	for _, child := range d.GetChildren() {
		children = append(children, toDeptInfo(child))
	}
	return &types.DeptInfo{
		Id:          d.GetId(),
		IdStr:       d.GetIdStr(),
		Name:        d.GetName(),
		ParentId:    d.GetParentId(),
		ParentIdStr: d.GetParentIdStr(),
		LeaderId:    d.GetLeaderId(),
		LeaderIdStr: d.GetLeaderIdStr(),
		LeaderName:  d.GetLeaderName(),
		Mobile:      d.GetMobile(),
		Email:       d.GetEmail(),
		Status:      d.GetStatus(),
		Sort:        d.GetSort(),
		TenantId:    d.GetTenantId(),
		TenantIdStr: d.GetTenantIdStr(),
		Children:    children,
		CreatedAt:   auditTime(d.GetCreatedAt()),
		CreatedBy:   d.GetCreatedBy(),
		UpdatedAt:   auditTime(d.GetUpdatedAt()),
		UpdatedBy:   d.GetUpdatedBy(),
	}
}

func toDeptInfoList(items []*apps.Dept) []*types.DeptInfo {
	out := make([]*types.DeptInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toDeptInfo(item))
	}
	return out
}

func toMenuInfo(m *apps.Menu) *types.MenuInfo {
	if m == nil {
		return nil
	}
	children := make([]*types.MenuInfo, 0, len(m.GetChildren()))
	for _, child := range m.GetChildren() {
		children = append(children, toMenuInfo(child))
	}
	return &types.MenuInfo{
		Id:          m.GetId(),
		IdStr:       m.GetIdStr(),
		MenuType:    m.GetMenuType(),
		Name:        m.GetName(),
		ParentId:    m.GetParentId(),
		ParentIdStr: m.GetParentIdStr(),
		Component:   m.GetComponent(),
		Path:        m.GetPath(),
		Icon:        m.GetIcon(),
		IsRedirect:  m.GetIsRedirect(),
		Redirect:    m.GetRedirect(),
		Hidden:      m.GetHidden(),
		Status:      m.GetStatus(),
		Sort:        m.GetSort(),
		Remark:      m.GetRemark(),
		Children:    children,
		CreatedAt:   auditTime(m.GetCreatedAt()),
		CreatedBy:   m.GetCreatedBy(),
		UpdatedAt:   auditTime(m.GetUpdatedAt()),
		UpdatedBy:   m.GetUpdatedBy(),
	}
}

func toMenuInfoList(items []*apps.Menu) []*types.MenuInfo {
	out := make([]*types.MenuInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toMenuInfo(item))
	}
	return out
}

func toDictDataInfo(d *apps.DictData) *types.DictDataInfo {
	if d == nil {
		return nil
	}
	return &types.DictDataInfo{
		Id:          d.GetId(),
		IdStr:       d.GetIdStr(),
		DictId:      d.GetDictId(),
		DictIdStr:   d.GetDictIdStr(),
		Name:        d.GetName(),
		Key:         d.GetKey(),
		Value:       d.GetValue(),
		Status:      d.GetStatus(),
		Remark:      d.GetRemark(),
		TenantId:    d.GetTenantId(),
		TenantIdStr: d.GetTenantIdStr(),
		CreatedAt:   auditTime(d.GetCreatedAt()),
		CreatedBy:   d.GetCreatedBy(),
		UpdatedAt:   auditTime(d.GetUpdatedAt()),
		UpdatedBy:   d.GetUpdatedBy(),
	}
}

func toDictDataInfoList(items []*apps.DictData) []*types.DictDataInfo {
	out := make([]*types.DictDataInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toDictDataInfo(item))
	}
	return out
}

func toDictInfo(d *apps.Dict) *types.DictInfo {
	if d == nil {
		return nil
	}
	return &types.DictInfo{
		Id:          d.GetId(),
		IdStr:       d.GetIdStr(),
		Name:        d.GetName(),
		Key:         d.GetKey(),
		Status:      d.GetStatus(),
		Remark:      d.GetRemark(),
		TenantId:    d.GetTenantId(),
		TenantIdStr: d.GetTenantIdStr(),
		DictData:    toDictDataInfoList(d.GetDictData()),
		CreatedAt:   auditTime(d.GetCreatedAt()),
		CreatedBy:   d.GetCreatedBy(),
		UpdatedAt:   auditTime(d.GetUpdatedAt()),
		UpdatedBy:   d.GetUpdatedBy(),
	}
}

func toDictInfoList(items []*apps.Dict) []*types.DictInfo {
	out := make([]*types.DictInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toDictInfo(item))
	}
	return out
}

func toTenantInfo(t *apps.Tenant) *types.TenantInfo {
	if t == nil {
		return nil
	}
	return &types.TenantInfo{
		Id:           t.GetId(),
		IdStr:        t.GetIdStr(),
		Name:         t.GetName(),
		Code:         t.GetCode(),
		AdminId:      t.GetAdminId(),
		AdminIdStr:   t.GetAdminIdStr(),
		AdminName:    t.GetAdminName(),
		ParentId:     t.GetParentId(),
		ParentIdStr:  t.GetParentIdStr(),
		PackageId:    t.GetPackageId(),
		PackageIdStr: t.GetPackageIdStr(),
		PackageName:  t.GetPackageName(),
		ExpiredAt:    auditTime(t.GetExpiredAt()),
		Status:       t.GetStatus(),
		Remark:       t.GetRemark(),
		CreatedAt:    auditTime(t.GetCreatedAt()),
		CreatedBy:    t.GetCreatedBy(),
		UpdatedAt:    auditTime(t.GetUpdatedAt()),
		UpdatedBy:    t.GetUpdatedBy(),
	}
}

func toTenantInfoList(items []*apps.Tenant) []*types.TenantInfo {
	out := make([]*types.TenantInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toTenantInfo(item))
	}
	return out
}

func toPackageInfo(p *apps.Package) *types.PackageInfo {
	if p == nil {
		return nil
	}
	return &types.PackageInfo{
		Id:        p.GetId(),
		IdStr:     p.GetIdStr(),
		Name:      p.GetName(),
		Code:      p.GetCode(),
		Status:    p.GetStatus(),
		Sort:      p.GetSort(),
		Remark:    p.GetRemark(),
		MenuIds:   idutil.ToStrings(p.GetMenuIds()),
		ApiIds:    idutil.ToStrings(p.GetApiIds()),
		CreatedAt: auditTime(p.GetCreatedAt()),
		CreatedBy: p.GetCreatedBy(),
		UpdatedAt: auditTime(p.GetUpdatedAt()),
		UpdatedBy: p.GetUpdatedBy(),
	}
}

func toPackageInfoList(items []*apps.Package) []*types.PackageInfo {
	out := make([]*types.PackageInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toPackageInfo(item))
	}
	return out
}

func toAPIInfo(a *apps.Api) *types.ApiInfo {
	if a == nil {
		return nil
	}
	return &types.ApiInfo{
		Id:        a.GetId(),
		IdStr:     a.GetIdStr(),
		ApiName:   a.GetApiName(),
		ApiType:   a.GetApiType(),
		ApiPath:   a.GetApiPath(),
		ApiMethod: a.GetApiMethod(),
		Status:    a.GetStatus(),
		Remark:    a.GetRemark(),
		CreatedAt: auditTime(a.GetCreatedAt()),
		CreatedBy: a.GetCreatedBy(),
		UpdatedAt: auditTime(a.GetUpdatedAt()),
		UpdatedBy: a.GetUpdatedBy(),
	}
}

func toAPIInfoList(items []*apps.Api) []*types.ApiInfo {
	out := make([]*types.ApiInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toAPIInfo(item))
	}
	return out
}
