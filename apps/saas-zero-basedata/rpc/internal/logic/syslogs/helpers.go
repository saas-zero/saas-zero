package syslogslogic

import (
	"github.com/saas-zero/saas-zero-common/pkg/id"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"google.golang.org/protobuf/proto"
)

func loginLogToResp(l *ent.SysLoginLog) *apps.LoginLog {
	return &apps.LoginLog{
		Id:       proto.Int64(l.ID),
		IdStr:    proto.String(id.ToString(l.ID)),
		Username: proto.String(l.Username),
		LoginIp:  proto.String(l.IP),
		Status:   proto.String(string(l.Status)),
		Msg:      proto.String(l.Message),
		LoginAt:  proto.Int64(l.LoginTime.UnixMilli()),
	}
}

func operationLogToResp(o *ent.SysOperationLog) *apps.OperationLog {
	return &apps.OperationLog{
		Id:            proto.Int64(o.ID),
		IdStr:         proto.String(id.ToString(o.ID)),
		Module:        proto.String(o.Module),
		Operation:     proto.String(o.Operation),
		RequestMethod: proto.String(o.Method),
		RequestUrl:    proto.String(o.Path),
		RequestParam:  proto.String(o.Params),
		ResponseData:  proto.String(o.Result),
		Duration:      proto.Int64(o.Duration),
		Status:        proto.String(o.Status),
		ErrorMsg:      proto.String(o.ErrorMsg),
		OperatorId:    proto.Int64(o.OperatorID),
		OperatorIdStr: proto.String(id.ToString(o.OperatorID)),
		OperatorName:  proto.String(o.OperatorName),
		OperatorIp:    proto.String(o.IP),
		TenantId:      proto.Int64(o.TenantID),
		TenantIdStr:   proto.String(id.ToString(o.TenantID)),
		CreatedAt:     proto.Int64(o.CreatedAt.UnixMilli()),
	}
}
