package syslogslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysoperationlog"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/pagination"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetOperationLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOperationLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOperationLogListLogic {
	return &GetOperationLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetOperationLogListLogic) GetOperationLogList(in *apps.LogPageReq) (*apps.OperationLogListResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)

	query := l.svcCtx.DB.SysOperationLog.Query().
		Where(sysoperationlog.TenantIDEQ(tenantId))

	if in.GetStatus() != "" {
		query = query.Where(sysoperationlog.Status(in.GetStatus()))
	}
	if in.GetModule() != "" {
		query = query.Where(sysoperationlog.ModuleContains(in.GetModule()))
	}
	if in.GetOperation() != "" {
		query = query.Where(sysoperationlog.OperationContains(in.GetOperation()))
	}
	if in.GetOperatorName() != "" {
		query = query.Where(sysoperationlog.OperatorNameContains(in.GetOperatorName()))
	}
	if in.GetPath() != "" {
		query = query.Where(sysoperationlog.PathContains(in.GetPath()))
	}

	total, err := query.Count(l.ctx)
	if err != nil {
		return nil, err
	}

	_, size, offset := pagination.Normalize(int(in.GetPage()), int(in.GetSize()))

	logs, err := query.
		Offset(offset).
		Limit(size).
		Order(ent.Desc(sysoperationlog.FieldID)).
		All(l.ctx)
	if err != nil {
		return nil, err
	}

	list := make([]*apps.OperationLog, len(logs))
	for i, log := range logs {
		list[i] = operationLogToResp(log)
	}
	return &apps.OperationLogListResp{
		Code:  int32(errno.Success.Code),
		Msg:   errno.Success.Msg,
		List:  list,
		Total: int64(total),
	}, nil
}
