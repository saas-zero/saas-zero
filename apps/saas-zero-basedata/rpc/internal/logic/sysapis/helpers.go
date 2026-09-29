package sysapislogic

import (
	"github.com/saas-zero/saas-zero-common/pkg/id"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"google.golang.org/protobuf/proto"
)

func apiToResp(a *ent.SysApi) *apps.Api {
	resp := &apps.Api{
		Id:        proto.Int64(a.ID),
		IdStr:     proto.String(id.ToString(a.ID)),
		ApiName:   proto.String(a.APIName),
		ApiType:   proto.String(string(a.APIType)),
		ApiPath:   proto.String(a.APIPath),
		ApiMethod: proto.String(string(a.APIMethod)),
		Status:    proto.String(string(a.Status)),
		Remark:    proto.String(a.Remark),
		CreatedAt: proto.Int64(a.CreatedAt.UnixMilli()),
		UpdatedAt: proto.Int64(a.UpdatedAt.UnixMilli()),
	}
	if a.CreatedBy != "" {
		resp.CreatedBy = proto.String(a.CreatedBy)
	}
	if a.UpdatedBy != "" {
		resp.UpdatedBy = proto.String(a.UpdatedBy)
	}
	return resp
}
