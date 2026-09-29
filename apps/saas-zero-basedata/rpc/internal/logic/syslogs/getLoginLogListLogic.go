package syslogslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysloginlog"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/pagination"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetLoginLogListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLoginLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLoginLogListLogic {
	return &GetLoginLogListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetLoginLogListLogic) GetLoginLogList(in *apps.LogPageReq) (*apps.LoginLogListResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)

	query := l.svcCtx.DB.SysLoginLog.Query().Where(sysloginlog.TenantIDEQ(tenantId))
	if in.GetUsername() != "" {
		query = query.Where(sysloginlog.UsernameContains(in.GetUsername()))
	}
	if in.GetStatus() != "" {
		query = query.Where(sysloginlog.StatusEQ(sysloginlog.Status(in.GetStatus())))
	}
	if in.GetIp() != "" {
		query = query.Where(sysloginlog.IPContains(in.GetIp()))
	}

	total, err := query.Count(l.ctx)
	if err != nil {
		return nil, err
	}

	_, size, offset := pagination.Normalize(int(in.GetPage()), int(in.GetSize()))

	logs, err := query.
		Offset(offset).
		Limit(size).
		Order(ent.Desc(sysloginlog.FieldLoginTime)).
		All(l.ctx)
	if err != nil {
		return nil, err
	}

	list := make([]*apps.LoginLog, len(logs))
	for i, log := range logs {
		list[i] = loginLogToResp(log)
	}
	return &apps.LoginLogListResp{
		Code:  int32(errno.Success.Code),
		Msg:   errno.Success.Msg,
		List:  list,
		Total: int64(total),
	}, nil
}
