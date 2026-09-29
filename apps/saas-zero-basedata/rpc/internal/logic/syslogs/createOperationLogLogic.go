package syslogslogic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
)

type CreateOperationLogLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOperationLogLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOperationLogLogic {
	return &CreateOperationLogLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateOperationLog 由 basedata-api 的操作日志中间件调用，写入一条操作审计日志。
// proto 的 requestMethod/requestUrl/requestParam/responseData/status/errorMsg/operatorIp
// 映射到 schema 的 method/path/params/result/status/error_msg/ip。
func (l *CreateOperationLogLogic) CreateOperationLog(in *apps.OperationLog) (*apps.EmptyResp, error) {
	result := in.GetResponseData()
	status := in.GetStatus()
	if status == "" {
		status = "success"
	}
	_, err := l.svcCtx.DB.SysOperationLog.Create().
		SetModule(in.GetModule()).
		SetOperation(in.GetOperation()).
		SetMethod(in.GetRequestMethod()).
		SetPath(in.GetRequestUrl()).
		SetParams(in.GetRequestParam()).
		SetResult(result).
		SetErrorMsg(in.GetErrorMsg()).
		SetStatus(status).
		SetDuration(in.GetDuration()).
		SetIP(in.GetOperatorIp()).
		SetOperatorID(in.GetOperatorId()).
		SetOperatorName(in.GetOperatorName()).
		SetTenantID(in.GetTenantId()).
		Save(l.ctx)
	if err != nil {
		return nil, err
	}
	return &apps.EmptyResp{Code: int32(errno.Success.Code), Msg: errno.Success.Msg}, nil
}
