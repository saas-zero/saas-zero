package systenantslogic

import (
	"github.com/saas-zero/saas-zero-common/pkg/id"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"google.golang.org/protobuf/proto"
)

func tenantToResp(t *ent.SysTenant) *apps.Tenant {
	resp := &apps.Tenant{
		Id:        proto.Int64(t.ID),
		IdStr:     proto.String(id.ToString(t.ID)),
		Name:      proto.String(t.Name),
		Code:      proto.String(t.Code),
		Status:    proto.String(string(t.Status)),
		Remark:    proto.String(t.Remark),
		CreatedAt: proto.Int64(t.CreatedAt.UnixMilli()),
		UpdatedAt: proto.Int64(t.UpdatedAt.UnixMilli()),
	}
	if t.CreatedBy != "" {
		resp.CreatedBy = proto.String(t.CreatedBy)
	}
	if t.UpdatedBy != "" {
		resp.UpdatedBy = proto.String(t.UpdatedBy)
	}
	if t.AdminID > 0 {
		resp.AdminId = proto.Int64(t.AdminID)
		resp.AdminIdStr = proto.String(id.ToString(t.AdminID))
	}
	if t.ParentID > 0 {
		resp.ParentId = proto.Int64(t.ParentID)
		resp.ParentIdStr = proto.String(id.ToString(t.ParentID))
	}
	if t.ParentName != "" {
		resp.ParentName = proto.String(t.ParentName)
	}
	if t.PackageID > 0 {
		resp.PackageId = proto.Int64(t.PackageID)
		resp.PackageIdStr = proto.String(id.ToString(t.PackageID))
	}
	if t.Edges.SysPackage != nil {
		resp.PackageName = proto.String(t.Edges.SysPackage.Name)
	}
	if !t.ExpiredAt.IsZero() {
		resp.ExpiredAt = proto.Int64(t.ExpiredAt.UnixMilli())
	}
	return resp
}
