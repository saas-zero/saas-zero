package sysdeptslogic

import (
	"github.com/saas-zero/saas-zero-common/pkg/id"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"google.golang.org/protobuf/proto"
)

func deptToResp(d *ent.SysDept) *apps.Dept {
	resp := &apps.Dept{
		Id:          proto.Int64(d.ID),
		IdStr:       proto.String(id.ToString(d.ID)),
		Name:        proto.String(d.Name),
		Status:      proto.String(string(d.Status)),
		Sort:        proto.Int32(int32(d.Sort)),
		Mobile:      proto.String(d.Mobile),
		Email:       proto.String(d.Email),
		TenantId:    proto.Int64(d.TenantID),
		TenantIdStr: proto.String(id.ToString(d.TenantID)),
		CreatedAt:   proto.Int64(d.CreatedAt.UnixMilli()),
		UpdatedAt:   proto.Int64(d.UpdatedAt.UnixMilli()),
	}
	if d.CreatedBy != "" {
		resp.CreatedBy = proto.String(d.CreatedBy)
	}
	if d.UpdatedBy != "" {
		resp.UpdatedBy = proto.String(d.UpdatedBy)
	}
	if d.ParentID > 0 {
		resp.ParentId = proto.Int64(d.ParentID)
		resp.ParentIdStr = proto.String(id.ToString(d.ParentID))
	}
	if d.ParentName != "" {
		resp.ParentName = proto.String(d.ParentName)
	}
	if d.LeaderID > 0 {
		resp.LeaderId = proto.Int64(d.LeaderID)
		resp.LeaderIdStr = proto.String(id.ToString(d.LeaderID))
	}
	if d.Edges.Leader != nil {
		resp.LeaderName = proto.String(d.Edges.Leader.Nickname)
	}
	return resp
}

func buildDeptTree(depts []*ent.SysDept, parentId int64) []*apps.Dept {
	var result []*apps.Dept
	for _, d := range depts {
		if d.ParentID == parentId {
			item := deptToResp(d)
			item.Children = buildDeptTree(depts, d.ID)
			result = append(result, item)
		}
	}
	return result
}
