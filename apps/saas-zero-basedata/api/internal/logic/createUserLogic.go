package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type CreateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateUserLogic) CreateUser(req *types.UserReq) (*types.BaseResp, error) {
	rpcReq := &apps.UserReq{
		Username: proto.String(req.Username),
		Password: proto.String(req.Password),
		Nickname: proto.String(req.Nickname),
		Mobile:   proto.String(req.Mobile),
		Email:    proto.String(req.Email),
		Status:   proto.String(req.Status),
	}
	if deptId := parseId(req.DeptId); deptId > 0 {
		rpcReq.DeptId = proto.Int64(deptId)
	}
	if req.Remark != "" {
		rpcReq.Remark = proto.String(req.Remark)
	}
	if len(req.RoleIds) > 0 {
		rpcReq.RoleIds = parseIds(req.RoleIds)
	}
	resp, err := l.svcCtx.SysUsers.CreateUser(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: resp.GetData()}, nil
}
