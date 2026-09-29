package sysapislogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysapi"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/pagination"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetApiListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetApiListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetApiListLogic {
	return &GetApiListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetApiListLogic) GetApiList(in *apps.ApiPageReq) (*apps.ApiListResp, error) {
	query := l.svcCtx.DB.SysApi.ActiveQuery()
	if in.GetApiName() != "" {
		query = query.Where(sysapi.APINameContains(in.GetApiName()))
	}
	if in.GetApiPath() != "" {
		query = query.Where(sysapi.APIPathContains(in.GetApiPath()))
	}
	if in.GetApiType() != "" {
		query = query.Where(sysapi.APITypeEQ(sysapi.APIType(in.GetApiType())))
	}
	if in.GetStatus() != "" {
		query = query.Where(sysapi.StatusEQ(sysapi.Status(in.GetStatus())))
	}

	total, err := query.Count(l.ctx)
	if err != nil {
		return nil, err
	}

	_, size, offset := pagination.Normalize(int(in.GetPage()), int(in.GetSize()))

	apis, err := query.
		Offset(offset).
		Limit(size).
		Order(ent.Asc(sysapi.FieldCreatedAt)).
		All(l.ctx)
	if err != nil {
		return nil, err
	}

	list := make([]*apps.Api, len(apis))
	for i, a := range apis {
		list[i] = apiToResp(a)
	}
	return &apps.ApiListResp{
		Code:  int32(errno.Success.Code),
		Msg:   errno.Success.Msg,
		List:  list,
		Total: int64(total),
	}, nil
}
