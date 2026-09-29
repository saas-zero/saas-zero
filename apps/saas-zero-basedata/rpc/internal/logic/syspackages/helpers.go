package syspackageslogic

import (
	"github.com/saas-zero/saas-zero-common/pkg/id"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"google.golang.org/protobuf/proto"
)

func packageToResp(p *ent.SysPackage) *apps.Package {
	resp := &apps.Package{
		Id:        proto.Int64(p.ID),
		IdStr:     proto.String(id.ToString(p.ID)),
		Name:      proto.String(p.Name),
		Code:      proto.String(p.Code),
		Status:    proto.String(string(p.Status)),
		Sort:      proto.Int32(int32(p.Sort)),
		Remark:    proto.String(p.Remark),
		CreatedAt: proto.Int64(p.CreatedAt.UnixMilli()),
		UpdatedAt: proto.Int64(p.UpdatedAt.UnixMilli()),
	}
	if p.CreatedBy != "" {
		resp.CreatedBy = proto.String(p.CreatedBy)
	}
	if p.UpdatedBy != "" {
		resp.UpdatedBy = proto.String(p.UpdatedBy)
	}
	if len(p.Edges.Menus) > 0 {
		menuIds := make([]int64, len(p.Edges.Menus))
		for i, m := range p.Edges.Menus {
			menuIds[i] = m.ID
		}
		resp.MenuIds = menuIds
	}
	if len(p.Edges.Apis) > 0 {
		apiIds := make([]int64, len(p.Edges.Apis))
		for i, a := range p.Edges.Apis {
			apiIds[i] = a.ID
		}
		resp.ApiIds = apiIds
	}
	return resp
}
