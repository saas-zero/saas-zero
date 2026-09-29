package sysapislogic

import (
	"context"
	"strconv"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysapi"
	"github.com/saas-zero/saas-zero-basedata/ent/sysrole"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/id"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyApisLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyApisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyApisLogic {
	return &GetMyApisLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetMyApisLogic) GetMyApis(_ *apps.EmptyReq) (*apps.ApiListResp, error) {
	userId := mixins.GetCurrentUserId(l.ctx)
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	isDefaultTenantAdmin := false
	var apis []*ent.SysApi
	var err error

	// 仅 default 租户的管理员（超级管理员）拥有全量 API；
	// 其他用户只能看到自己角色（在 Casbin 策略中）被分配的 API 及其 group 父节点
	if userId > 0 {
		user, uerr := l.svcCtx.DB.SysUser.ActiveQuery().
			Where(sysuser.IDEQ(userId)).
			WithRoles(func(q *ent.SysRoleQuery) {
				q.Where(sysrole.DeletedAtIsNil(), sysrole.StatusEQ(sysrole.StatusActive))
			}).
			Only(l.ctx)
		if uerr == nil {
			roleCodes := make([]string, 0, len(user.Edges.Roles))
			hasAdminRole := false
			for _, role := range user.Edges.Roles {
				if role.Code == "admin" {
					hasAdminRole = true
				}
				roleCodes = append(roleCodes, role.Code)
			}

			if hasAdminRole && tenantId > 0 {
				tenant, terr := l.svcCtx.DB.SysTenant.Get(l.ctx, tenantId)
				if terr == nil && tenant.Code == "default" {
					isDefaultTenantAdmin = true
				}
			}

			if isDefaultTenantAdmin {
				apis, err = l.svcCtx.DB.SysApi.ActiveQuery().
					Order(ent.Asc(sysapi.FieldCreatedAt)).
					All(l.ctx)
				if err != nil {
					return nil, err
				}
			} else {
				apis, err = l.myRoleApis(roleCodes, tenantId)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	list := make([]*apps.Api, len(apis))
	for i, a := range apis {
		list[i] = apiToResp(a)
	}
	return &apps.ApiListResp{
		Code:  int32(errno.Success.Code),
		Msg:   errno.Success.Msg,
		List:  list,
		Total: int64(len(list)),
	}, nil
}

// myRoleApis 返回当前用户全部角色在 Casbin 策略中被分配的 API，
// 并补上这些 API 所属的 group 分组节点，避免前端树出现悬空的接口。
func (l *GetMyApisLogic) myRoleApis(roleCodes []string, tenantId int64) ([]*ent.SysApi, error) {
	if len(roleCodes) == 0 || l.svcCtx.Enforcer == nil {
		return nil, nil
	}
	dom := id.ToString(tenantId)
	apiIdSet := make(map[int64]bool)
	for _, code := range roleCodes {
		policies, _ := l.svcCtx.Enforcer.GetFilteredPolicy(0, code, dom)
		for _, p := range policies {
			if len(p) <= 4 {
				continue
			}
			if id, err := strconv.ParseInt(p[4], 10, 64); err == nil {
				apiIdSet[id] = true
			}
		}
	}
	if len(apiIdSet) == 0 {
		return nil, nil
	}

	all, err := l.svcCtx.DB.SysApi.ActiveQuery().All(l.ctx)
	if err != nil {
		return nil, err
	}
	byPath := make(map[string]*ent.SysApi)
	for _, a := range all {
		byPath[a.APIPath] = a
	}

	final := make(map[int64]bool)
	for _, a := range all {
		if !apiIdSet[a.ID] {
			continue
		}
		final[a.ID] = true
		// 通过路径前缀补齐所属 group
		for path, g := range byPath {
			if g.APIType == sysapi.APITypeGroup && path != "" && path+"/" != a.APIPath && len(a.APIPath) > len(path) && a.APIPath[:len(path)] == path && a.APIPath[len(path)] == '/' {
				final[g.ID] = true
				break
			}
		}
	}

	ids := make([]int64, 0, len(final))
	for id := range final {
		ids = append(ids, id)
	}
	return l.svcCtx.DB.SysApi.ActiveQuery().
		Where(sysapi.IDIn(ids...)).
		Order(ent.Asc(sysapi.FieldCreatedAt)).
		All(l.ctx)
}