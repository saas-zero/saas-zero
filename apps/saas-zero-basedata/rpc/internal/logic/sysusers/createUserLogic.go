package sysuserslogic

import (
	"context"
	"github.com/saas-zero/saas-zero-common/pkg/id"

	"github.com/saas-zero/saas-zero-basedata/ent/sysdept"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/proto"
)

type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateUserLogic) CreateUser(in *apps.UserReq) (*apps.UserResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)

	ctx := mixins.SetCurrentTenantId(l.ctx, tenantId)
	ctx = mixins.SetCurrentUserId(ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	if err := checkUserRolesInTenant(l.svcCtx, ctx, tenantId, in.GetRoleIds()); err != nil {
		return nil, errno.New(errno.InvalidParam.Code, err.Error())
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 同租户下用户名唯一：预检查给出友好错误，数据库 (tenant_id, username) 部分唯一索引兜底
	exists, err := l.svcCtx.DB.SysUser.TenantQuery(tenantId).
		Where(sysuser.UsernameEQ(in.GetUsername())).
		Exist(ctx)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errno.UsernameExists
	}

	create := l.svcCtx.DB.SysUser.Create().
		SetUsername(in.GetUsername()).
		SetPassword(string(hash)).
		SetNickname(in.GetNickname()).
		SetMobile(in.GetMobile()).
		SetEmail(in.GetEmail()).
		SetStatus(sysuser.Status(in.GetStatus()))

	if in.GetDeptId() > 0 {
		deptInTenant, derr := l.svcCtx.DB.SysDept.TenantQuery(tenantId).
			Where(sysdept.IDEQ(in.GetDeptId())).
			Exist(ctx)
		if derr != nil {
			return nil, derr
		}
		if !deptInTenant {
			return nil, errno.New(errno.InvalidParam.Code, "部门不存在或不属于当前租户")
		}
		create.SetDeptID(in.GetDeptId())
	}
	if in.GetRemark() != "" {
		create.SetRemark(in.GetRemark())
	}

	result, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	if len(in.GetRoleIds()) > 0 {
		if err := l.svcCtx.DB.SysUser.UpdateOneID(result.ID).
			AddRoleIDs(in.GetRoleIds()...).
			Exec(ctx); err != nil {
			return nil, err
		}
	}

	return &apps.UserResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: &apps.User{
			Id:     proto.Int64(result.ID),
			IdStr:  proto.String(id.ToString(result.ID)),
			Status: proto.String(string(result.Status)),
		},
	}, nil
}
