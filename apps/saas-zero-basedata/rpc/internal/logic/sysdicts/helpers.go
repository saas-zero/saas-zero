package sysdictslogic

import (
	"github.com/saas-zero/saas-zero-common/pkg/id"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"google.golang.org/protobuf/proto"
)

func dictToResp(d *ent.SysDict) *apps.Dict {
	resp := &apps.Dict{
		Id:          proto.Int64(d.ID),
		IdStr:       proto.String(id.ToString(d.ID)),
		Name:        proto.String(d.Name),
		Key:         proto.String(d.Key),
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
	if len(d.Edges.SysDictDatas) > 0 {
		list := make([]*apps.DictData, len(d.Edges.SysDictDatas))
		for i, dd := range d.Edges.SysDictDatas {
			list[i] = dictDataToResp(dd)
		}
		resp.DictData = list
	}
	return resp
}

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
