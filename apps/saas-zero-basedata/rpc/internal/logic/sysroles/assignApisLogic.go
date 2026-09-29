package sysroleslogic

import (
	"context"
	"fmt"
	"github.com/saas-zero/saas-zero-common/pkg/id"
	"github.com/zeromicro/go-zero/core/logx"
	"strings"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysapi"
	"github.com/saas-zero/saas-zero-basedata/ent/sysrole"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
)

type AssignApisLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAssignApisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AssignApisLogic {
	return &AssignApisLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AssignApisLogic) AssignApis(in *apps.RoleReq) (*apps.EmptyResp, error) {
	roleCode := in.GetCode()
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	dom := id.ToString(tenantId)

	// 前端只传 id + apiIds，不传 code，需从角色表查询兜底，
	// 否则会生成 v0 为空的脏策略。同时校验系统内置角色不可改权限。
	// 目标角色必须属于当前租户，防止跨租户操作。
	role, err := l.svcCtx.DB.SysRole.TenantQuery(tenantId).
		Where(sysrole.IDEQ(in.GetId())).
		Only(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errno.New(errno.InvalidParam.Code, "角色不存在或不属于当前租户")
		}
		return nil, err
	}
	if role.IsSystem {
		return nil, errno.New(errno.InvalidParam.Code, fmt.Sprintf("系统内置角色「%s」不可修改", role.Name))
	}
	if roleCode == "" {
		roleCode = role.Code
	} else if roleCode != role.Code {
		return nil, errno.New(errno.InvalidParam.Code, "角色编码与角色 ID 不匹配")
	}

	// 继承式授权：只能把当前用户自己拥有的 API 授给别人
	if err := checkAssignableApis(l.svcCtx, l.ctx, in.GetApiIds()); err != nil {
		return nil, err
	}

	if l.svcCtx.Enforcer == nil {
		return nil, errno.AuthServiceUnavailable
	}
	// API assignment is an edit of the role even though the association is
	// stored in Casbin. Touch the role so its audit fields stay authoritative.
	if err := l.svcCtx.DB.SysRole.UpdateOne(role).Exec(l.ctx); err != nil {
		return nil, err
	}

	if _, err := l.svcCtx.Enforcer.RemoveFilteredPolicy(0, roleCode, dom); err != nil {
		return nil, err
	}

	for _, apiId := range in.GetApiIds() {
		api, err := l.svcCtx.DB.SysApi.ActiveQuery().
			Where(sysapi.IDEQ(apiId)).
			Only(l.ctx)
		if err != nil {
			return nil, err
		}
		// 只对具体接口（api 类型）生成策略；目录（group）仅用于分组展示，不参与权限匹配
		if api.APIType != sysapi.APITypeAPI {
			continue
		}
		if _, err := l.svcCtx.Enforcer.AddPolicy(roleCode, dom, api.APIPath, strings.ToUpper(string(api.APIMethod)), id.ToString(apiId)); err != nil {
			return nil, err
		}
	}

	users, err := l.svcCtx.DB.SysUser.Query().
		Where(sysuser.TenantIDEQ(tenantId), sysuser.DeletedAtIsNil()).
		Where(sysuser.HasRolesWith(sysrole.CodeEQ(roleCode), sysrole.DeletedAtIsNil())).
		All(l.ctx)
	if err == nil {
		// API 权限变更后，旧 token 对应的策略已不一致，必须踢掉旧会话
		if err := svc.BumpUsersTokenVersion(l.svcCtx.Redis, users); err != nil {
			return nil, err
		}
	}

	return &apps.EmptyResp{Code: int32(errno.Success.Code), Msg: errno.Success.Msg}, nil
}
