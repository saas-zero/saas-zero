package logic

import (
	"context"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type GetDictDataByDictKeyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetDictDataByDictKeyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDictDataByDictKeyLogic {
	return &GetDictDataByDictKeyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDictDataByDictKeyLogic) GetDictDataByDictKey(req *types.DictReq) (*types.BaseResp, error) {
	resp, err := l.svcCtx.SysDictDatas.GetDictDataByDictKey(l.ctx, &apps.DictReq{Key: proto.String(req.Key)})
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{
		Code: int(resp.Code),
		Msg:  resp.Msg,
		Data: &types.PageResp{List: toDictDataInfoList(resp.List), Total: resp.Total},
	}, nil
}
