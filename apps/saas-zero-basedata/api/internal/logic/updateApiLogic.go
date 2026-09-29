package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type UpdateApiLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateApiLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateApiLogic {
	return &UpdateApiLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateApiLogic) UpdateApi(req *types.ApiReq) (*types.BaseResp, error) {
	rpcReq := &apps.ApiReq{Id: proto.Int64(parseId(req.Id))}
	if req.ApiName != "" {
		rpcReq.ApiName = proto.String(req.ApiName)
	}
	if req.ApiPath != "" {
		rpcReq.ApiPath = proto.String(req.ApiPath)
	}
	if req.ApiMethod != "" {
		rpcReq.ApiMethod = proto.String(req.ApiMethod)
	}
	if req.ApiType != "" {
		rpcReq.ApiType = proto.String(req.ApiType)
	}
	if req.Status != "" {
		rpcReq.Status = proto.String(req.Status)
	}
	if req.Remark != "" {
		rpcReq.Remark = proto.String(req.Remark)
	}
	resp, err := l.svcCtx.SysApis.UpdateApi(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: resp.GetData()}, nil
}
