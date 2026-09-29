package sysuserslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type GetUserByUsernameLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserByUsernameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserByUsernameLogic {
	return &GetUserByUsernameLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserByUsernameLogic) GetUserByUsername(in *apps.UserReq) (*apps.UserResp, error) {
	tenantId := mixins.GetCurrentTenantId(l.ctx)

	u, err := l.svcCtx.DB.SysUser.TenantQuery(tenantId).
		Where(sysuser.UsernameEQ(in.GetUsername())).
		WithRoles().
		WithSysDept().
		Only(l.ctx)
	if err != nil {
		return nil, err
	}
	resp := userToResp(u)
	// Password is only returned via GetUserByUsername (login flow) for bcrypt verification.
	// All other user queries intentionally exclude it for security.
	resp.Password = proto.String(u.Password)
	// Lockout state is only exposed on the login flow so auth can pre-check
	// whether the account is currently locked before verifying the password.
	resp.LoginErrorCount = proto.Int32(u.LoginErrorCount)
	if !u.LockoutUntil.IsZero() {
		resp.LockoutUntil = proto.Int64(u.LockoutUntil.UnixMilli())
	}
	return &apps.UserResp{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Msg,
		Data: resp,
	}, nil
}
