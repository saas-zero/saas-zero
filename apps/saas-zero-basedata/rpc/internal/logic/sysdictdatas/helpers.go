package sysdictdataslogic

import (
	"github.com/saas-zero/saas-zero-common/pkg/id"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"google.golang.org/protobuf/proto"
)

func dictDataToResp(d *ent.SysDictData) *apps.DictData {
	resp := &apps.DictData{
		Id:          proto.Int64(d.ID),
		IdStr:       proto.String(id.ToString(d.ID)),
		DictId:      proto.Int64(d.DictID),
		DictIdStr:   proto.String(id.ToString(d.DictID)),
		Name:        proto.String(d.Name),
		Key:         proto.String(d.Key),
		Value:       proto.String(d.Value),
		Status:      proto.String(string(d.Status)),
		Remark:      proto.String(d.Remark),
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
	return resp
}
