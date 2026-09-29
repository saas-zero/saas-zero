package sysdictslogic

import (
	"context"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/id"

	"github.com/saas-zero/saas-zero-basedata/ent/sysdict"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type CreateDictLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateDictLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDictLogic {
	return &CreateDictLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateDictLogic) CreateDict(in *apps.DictReq) (*apps.DictResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)

	ctx := mixins.SetCurrentTenantId(l.ctx, tenantId)
	ctx = mixins.SetCurrentUserId(ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	create := l.svcCtx.DB.SysDict.Create().
		SetName(in.GetName()).
		SetKey(in.GetKey()).
		SetStatus(sysdict.Status(in.GetStatus()))

	if in.GetRemark() != "" {
		create.SetRemark(in.GetRemark())
	}

	result, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.DictResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: &apps.Dict{
			Id:    proto.Int64(result.ID),
			IdStr: proto.String(id.ToString(result.ID)),
		},
	}, nil
}
