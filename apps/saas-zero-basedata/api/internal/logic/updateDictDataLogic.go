package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type UpdateDictDataLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateDictDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDictDataLogic {
	return &UpdateDictDataLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateDictDataLogic) UpdateDictData(req *types.DictDataReq) (*types.BaseResp, error) {
	rpcReq := &apps.DictDataReq{Id: proto.Int64(parseId(req.Id))}
	if req.DictId != "" {
		rpcReq.DictId = proto.Int64(parseId(req.DictId))
	}
	if req.Name != "" {
		rpcReq.Name = proto.String(req.Name)
	}
	if req.Key != "" {
		rpcReq.Key = proto.String(req.Key)
	}
	if req.Value != "" {
		rpcReq.Value = proto.String(req.Value)
	}
	if req.Status != "" {
		rpcReq.Status = proto.String(req.Status)
	}
	if req.Remark != "" {
		rpcReq.Remark = proto.String(req.Remark)
	}
	resp, err := l.svcCtx.SysDictDatas.UpdateDictData(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: resp.GetData()}, nil
}
