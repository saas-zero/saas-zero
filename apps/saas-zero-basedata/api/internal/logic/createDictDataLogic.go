package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type CreateDictDataLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateDictDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateDictDataLogic {
	return &CreateDictDataLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateDictDataLogic) CreateDictData(req *types.DictDataReq) (*types.BaseResp, error) {
	rpcReq := &apps.DictDataReq{
		DictId: proto.Int64(parseId(req.DictId)),
		Name:   proto.String(req.Name),
		Key:    proto.String(req.Key),
		Value:  proto.String(req.Value),
		Status: proto.String(req.Status),
	}
	if req.Remark != "" {
		rpcReq.Remark = proto.String(req.Remark)
	}
	resp, err := l.svcCtx.SysDictDatas.CreateDictData(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: resp.GetData()}, nil
}
