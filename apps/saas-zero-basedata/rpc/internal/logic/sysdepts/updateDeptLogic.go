package sysdeptslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysdept"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDeptLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDeptLogic {
	return &UpdateDeptLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateDeptLogic) UpdateDept(in *apps.DeptReq) (*apps.DeptResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)
	userId := mixins.GetCurrentUserId(l.ctx)
	userName := mixins.GetCurrentUserName(l.ctx)
	ctx := mixins.SetCurrentTenantId(l.ctx, tenantId)
	ctx = mixins.SetCurrentUserId(ctx, userId)
	ctx = mixins.SetCurrentUserName(ctx, userName)

	// 目标部门必须属于当前租户，防止跨租户按全局 ID 更新
	target, err := l.svcCtx.DB.SysDept.TenantQuery(tenantId).
		Where(sysdept.IDEQ(in.GetId())).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errno.New(errno.InvalidParam.Code, "部门不存在或不属于当前租户")
		}
		return nil, err
	}

	update := l.svcCtx.DB.SysDept.UpdateOne(target)
	if in.Name != nil {
		update.SetName(in.GetName())
	}
	if in.ParentId != nil {
		if in.GetParentId() > 0 {
			// 父部门必须属于当前租户
			if in.GetParentId() == target.ID {
				return nil, errno.New(errno.InvalidParam.Code, "不能将部门自身设为父部门")
			}
			parent, err := l.svcCtx.DB.SysDept.TenantQuery(tenantId).
				Where(sysdept.IDEQ(in.GetParentId())).
				Only(ctx)
			if err != nil {
				if ent.IsNotFound(err) {
					return nil, errno.New(errno.InvalidParam.Code, "父部门不存在或不属于当前租户")
				}
				return nil, err
			}
			update.SetParentID(in.GetParentId()).SetParentName(parent.Name)
		} else {
			// 移除父级：清空 parent_id 与冗余 parent_name
			update.ClearParentID().SetParentName("")
		}
	}
	if in.LeaderId != nil {
		if in.GetLeaderId() > 0 {
			// 负责人必须属于当前租户
			leaderInTenant, lerr := l.svcCtx.DB.SysUser.TenantQuery(tenantId).
				Where(sysuser.IDEQ(in.GetLeaderId())).
				Exist(ctx)
			if lerr != nil {
				return nil, lerr
			}
			if !leaderInTenant {
				return nil, errno.New(errno.InvalidParam.Code, "部门负责人不存在或不属于当前租户")
			}
			update.SetLeaderID(in.GetLeaderId())
		} else {
			update.ClearLeaderID()
		}
	}
	if in.Mobile != nil {
		update.SetMobile(in.GetMobile())
	}
	if in.Email != nil {
		update.SetEmail(in.GetEmail())
	}
	if in.Status != nil {
		update.SetStatus(sysdept.Status(in.GetStatus()))
	}
	if in.Sort != nil {
		update.SetSort(uint32(in.GetSort()))
	}

	result, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	d, err := l.svcCtx.DB.SysDept.TenantQuery(tenantId).Where(sysdept.IDEQ(result.ID)).WithLeader().Only(ctx)
	if err != nil {
		return nil, err
	}
	return &apps.DeptResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: deptToResp(d),
	}, nil
}
