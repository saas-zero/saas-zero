package sysdictdataslogic

import (
	"context"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/id"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdict"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdictdata"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/logic/tenantcheck"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type CreateDictDataLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateDictDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDictDataLogic {
	return &CreateDictDataLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateDictDataLogic) CreateDictData(in *apps.DictDataReq) (*apps.DictDataResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)

	ctx := mixins.SetCurrentTenantId(l.ctx, tenantId)
	ctx = mixins.SetCurrentUserId(ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	// 所属字典必须对当前租户可见（系统默认或被授权范围），且系统字典仅系统管理员可新增数据
	dict, err := l.svcCtx.DB.SysDict.TenantAwareQuery(tenantId).
		Where(sysdict.IDEQ(in.GetDictId())).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errno.New(errno.InvalidParam.Code, "字典不存在")
		}
		return nil, err
	}
	if dict.TenantID == 0 {
		if err := tenantcheck.RequireDefaultTenantAdmin(l.svcCtx, ctx); err != nil {
			return nil, err
		}
	}

	create := l.svcCtx.DB.SysDictData.Create().
		SetDictID(in.GetDictId()).
		SetName(in.GetName()).
		SetKey(in.GetKey()).
		SetValue(in.GetValue()).
		SetStatus(sysdictdata.Status(in.GetStatus()))

	if in.GetRemark() != "" {
		create.SetRemark(in.GetRemark())
	}

	result, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.DictDataResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: &apps.DictData{
			Id:    proto.Int64(result.ID),
			IdStr: proto.String(id.ToString(result.ID)),
		},
	}, nil
}
