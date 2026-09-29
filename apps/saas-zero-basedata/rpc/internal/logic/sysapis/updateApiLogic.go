package sysapislogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent/sysapi"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateApiLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateApiLogic {
	return &UpdateApiLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateApiLogic) UpdateApi(in *apps.ApiReq) (*apps.ApiResp, error) {
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentUserId(l.ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	update := l.svcCtx.DB.SysApi.UpdateOneID(in.GetId())
	if in.ApiName != nil {
		update.SetAPIName(in.GetApiName())
	}
	if in.ApiType != nil {
		update.SetAPIType(sysapi.APIType(in.GetApiType()))
	}
	if in.ApiPath != nil {
		update.SetAPIPath(in.GetApiPath())
	}
	if in.ApiMethod != nil {
		update.SetAPIMethod(sysapi.APIMethod(in.GetApiMethod()))
	}
	if in.Status != nil {
		update.SetStatus(sysapi.Status(in.GetStatus()))
	}
	if in.Remark != nil {
		update.SetRemark(in.GetRemark())
	}

	result, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	a, err := l.svcCtx.DB.SysApi.Query().Where(sysapi.IDEQ(result.ID)).Only(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.ApiResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: apiToResp(a),
	}, nil
}
