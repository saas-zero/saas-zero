package systenantslogic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"

	idutil "github.com/saas-zero/saas-zero-common/pkg/id"
)

type GetTenantUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTenantUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantUsersLogic {
	return &GetTenantUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetTenantUsers 返回指定租户下的用户列表（供编辑租户时选择管理员）。
func (l *GetTenantUsersLogic) GetTenantUsers(in *apps.TenantReq) (*apps.UserListResp, error) {
	users, err := l.svcCtx.DB.SysUser.ActiveQuery().
		Where(sysuser.TenantIDEQ(in.GetId())).
		Order(ent.Asc(sysuser.FieldUsername)).
		All(l.ctx)
	if err != nil {
		return nil, err
	}

	list := make([]*apps.User, 0, len(users))
	for _, u := range users {
		list = append(list, &apps.User{
			Id:       proto.Int64(u.ID),
			IdStr:    proto.String(idutil.ToString(u.ID)),
			Username: proto.String(u.Username),
			Nickname: proto.String(u.Nickname),
		})
	}
	return &apps.UserListResp{
		Code:  int32(errno.Success.Code),
		Msg:   errno.Success.Msg,
		List:  list,
		Total: int64(len(list)),
	}, nil
}
