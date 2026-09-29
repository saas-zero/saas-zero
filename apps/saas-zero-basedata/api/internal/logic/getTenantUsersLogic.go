package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type GetTenantUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantUsersLogic {
	return &GetTenantUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetTenantUsers 返回指定租户下的用户（供编辑租户时选择管理员）。
func (l *GetTenantUsersLogic) GetTenantUsers(req *types.TenantReq) (*types.BaseResp, error) {
	resp, err := l.svcCtx.SysTenants.GetTenantUsers(l.ctx, &apps.TenantReq{
		Id: proto.Int64(parseId(req.Id)),
	})
	if err != nil {
		return nil, err
	}
	// 仅返回选择管理员需要的字段
	list := make([]map[string]interface{}, 0, len(resp.List))
	for _, u := range resp.List {
		list = append(list, map[string]interface{}{
			"idStr":    u.GetIdStr(),
			"username": u.GetUsername(),
			"nickname": u.GetNickname(),
		})
	}
	return &types.BaseResp{
		Code: int(resp.Code),
		Msg:  resp.Msg,
		Data: map[string]interface{}{"list": list},
	}, nil
}
