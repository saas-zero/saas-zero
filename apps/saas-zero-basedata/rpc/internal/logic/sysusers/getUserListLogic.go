package sysuserslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-common/pkg/pagination"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserListLogic {
	return &GetUserListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserListLogic) GetUserList(in *apps.UserPageReq) (*apps.UserListResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)

	query := l.svcCtx.DB.SysUser.TenantQuery(tenantId)
	if in.GetUsername() != "" {
		query = query.Where(sysuser.UsernameContains(in.GetUsername()))
	}
	if in.GetNickname() != "" {
		query = query.Where(sysuser.NicknameContains(in.GetNickname()))
	}
	if in.GetMobile() != "" {
		query = query.Where(sysuser.MobileContains(in.GetMobile()))
	}
	if in.GetStatus() != "" {
		query = query.Where(sysuser.StatusEQ(sysuser.Status(in.GetStatus())))
	}
	if in.GetDeptId() > 0 {
		query = query.Where(sysuser.DeptIDEQ(in.GetDeptId()))
	}

	total, err := query.Count(l.ctx)
	if err != nil {
		logx.Errorf("GetUserList gRPC error: %v", err)
		return nil, err
	}

	_, size, offset := pagination.Normalize(int(in.GetPage()), int(in.GetSize()))

	users, err := query.
		Offset(offset).
		Limit(size).
		Order(ent.Asc(sysuser.FieldCreatedAt)).
		WithRoles().
		WithSysDept().
		All(l.ctx)
	if err != nil {
		return nil, err
	}

	list := make([]*apps.User, len(users))
	for i, u := range users {
		list[i] = userToResp(u)
	}

	return &apps.UserListResp{
		Code:  int32(errno.Success.Code),
		Msg:   errno.Success.Msg,
		List:  list,
		Total: int64(total),
	}, nil
}
