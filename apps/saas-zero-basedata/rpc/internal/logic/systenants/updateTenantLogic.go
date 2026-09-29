package systenantslogic

import (
	"context"
	"strings"
	"time"

	"github.com/saas-zero/saas-zero-basedata/ent/syspackage"
	"github.com/saas-zero/saas-zero-basedata/ent/sysrole"
	"github.com/saas-zero/saas-zero-basedata/ent/systenant"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	idutil "github.com/saas-zero/saas-zero-common/pkg/id"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTenantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantLogic {
	return &UpdateTenantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateTenantLogic) UpdateTenant(in *apps.TenantReq) (*apps.TenantResp, error) {
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentUserId(l.ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	oldTenant, err := l.svcCtx.DB.SysTenant.Get(ctx, in.GetId())
	if err != nil {
		return nil, err
	}

	update := l.svcCtx.DB.SysTenant.UpdateOneID(in.GetId())
	if in.Name != nil {
		update.SetName(in.GetName())
	}
	if in.Code != nil {
		update.SetCode(in.GetCode())
	}
	if in.AdminId != nil {
		update.SetAdminID(in.GetAdminId())
	}
	if in.ParentId != nil {
		if in.GetParentId() > 0 {
			parent, err := l.svcCtx.DB.SysTenant.Query().Where(systenant.IDEQ(in.GetParentId())).Only(ctx)
			if err != nil {
				return nil, err
			}
			update.SetParentID(in.GetParentId()).SetParentName(parent.Name)
		} else {
			// 移除父级：清空 parent_id 与冗余 parent_name
			update.ClearParentID().SetParentName("")
		}
	}
	if in.PackageId != nil && in.GetPackageId() > 0 {
		update.SetPackageID(in.GetPackageId())
	}
	if in.ExpiredAt != nil {
		update.SetExpiredAt(time.UnixMilli(in.GetExpiredAt()))
	}
	if in.Status != nil {
		update.SetStatus(systenant.Status(in.GetStatus()))
	}
	if in.Remark != nil {
		update.SetRemark(in.GetRemark())
	}

	result, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	// 换套餐：把该租户 admin 角色的菜单/API 权限重新对齐到新套餐，并踢掉相关会话
	if in.PackageId != nil && in.GetPackageId() > 0 && in.GetPackageId() != oldTenant.PackageID {
		if err := l.syncAdminRoleToPackage(ctx, result.ID, in.GetPackageId()); err != nil {
			return nil, err
		}
	}

	t, err := l.svcCtx.DB.SysTenant.Query().Where(systenant.IDEQ(result.ID)).WithSysPackage().Only(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.TenantResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: tenantToResp(t),
	}, nil
}

// syncAdminRoleToPackage 换套餐后重新对齐 admin 角色权限：
//  1. 清除 admin 角色的菜单关联，重新继承新套餐菜单
//  2. 重建 Casbin 策略（新套餐 API）
//  3. 递增拥有 admin 角色的用户 token_version，强制重新登录获取新权限
func (l *UpdateTenantLogic) syncAdminRoleToPackage(ctx context.Context, tenantId, newPackageId int64) error {
	pkg, err := l.svcCtx.DB.SysPackage.Query().
		Where(syspackage.IDEQ(newPackageId)).
		WithMenus().WithApis().
		Only(ctx)
	if err != nil {
		return err
	}
	menuIDs := make([]int64, 0, len(pkg.Edges.Menus))
	for _, m := range pkg.Edges.Menus {
		menuIDs = append(menuIDs, m.ID)
	}

	role, err := l.svcCtx.DB.SysRole.Query().
		Where(sysrole.TenantIDEQ(tenantId), sysrole.CodeEQ("admin"), sysrole.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		// 没有 admin 角色（异常租户）不阻塞换套餐
		return nil
	}

	if err := l.svcCtx.DB.SysRole.UpdateOneID(role.ID).
		ClearMenus().
		AddMenuIDs(menuIDs...).
		Exec(ctx); err != nil {
		return err
	}

	// 重建 Casbin：角色 code 未变（admin），需克"admin"策略、重加新套餐 API
	dom := idutil.ToString(tenantId)
	if l.svcCtx.Enforcer != nil {
		l.svcCtx.Enforcer.RemoveFilteredPolicy(0, role.Code, dom)
		for _, api := range pkg.Edges.Apis {
			if _, err := l.svcCtx.Enforcer.AddPolicy(role.Code, dom, api.APIPath,
				strings.ToUpper(string(api.APIMethod)), idutil.ToString(api.ID)); err != nil {
				logx.Errorf("updateTenant: add casbin policy failed: %v", err)
			}
		}
	}

	// 踢掉 admin 角色用户的旧会话
	users, err := l.svcCtx.DB.SysUser.Query().
		Where(sysuser.HasRolesWith(sysrole.IDEQ(role.ID))).
		All(ctx)
	if err == nil {
		// 换套餐后 admin 权限已变，旧会话必须失效（递增失败不能当成成功）
		if err := svc.BumpUsersTokenVersion(l.svcCtx.Redis, users); err != nil {
			return err
		}
	}
	return nil
}
