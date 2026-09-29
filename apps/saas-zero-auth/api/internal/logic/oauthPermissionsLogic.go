// Code scaffolded by goctl. Safe to edit.

package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-auth/api/internal/svc"
	"github.com/saas-zero/saas-zero-auth/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

type OauthPermissionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOauthPermissionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OauthPermissionsLogic {
	return &OauthPermissionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OauthPermissionsLogic) OauthPermissions() (resp *types.BaseResp, err error) {
	// 统一会话校验：签名 + 有效期 + Redis JTI + tokenVersion
	_, ok := validateSession(l.svcCtx.Redis, l.svcCtx.Config.JwtSecret, GetToken(l.ctx))
	if ok != nil {
		return &types.BaseResp{Code: ok.Code, Msg: ok.Msg}, nil
	}
	treeResp, err := l.svcCtx.SysMenus.GetMenuTree(withAuthContext(l.ctx, l.svcCtx.Config.JwtSecret), &apps.EmptyReq{})
	if err != nil {
		return nil, err
	}
	// 权限码来自 button 类型菜单的 path 字段（如 "system:user:create"）
	perms := collectButtonPerms(treeResp.GetData())
	if perms == nil {
		perms = []string{}
	}
	return &types.BaseResp{
		Code: errno.Success.Code, Msg: errno.Success.Msg, Data: perms,
	}, nil
}

func collectButtonPerms(menus []*apps.Menu) []string {
	perms := []string{}
	for _, m := range menus {
		if m.GetMenuType() == "button" && m.GetPath() != "" {
			perms = append(perms, m.GetPath())
		}
		perms = append(perms, collectButtonPerms(m.GetChildren())...)
	}
	return perms
}
