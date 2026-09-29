package logic

import (
	"context"
	"encoding/json"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-common/pkg/timex"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type GetOperationLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetOperationLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOperationLogListLogic {
	return &GetOperationLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetOperationLogListLogic) GetOperationLogList(req *types.LogPageReq) (*types.BaseResp, error) {
	rpcReq := &apps.LogPageReq{
		Page: int32(req.Page),
		Size: int32(req.PageSize),
	}
	if req.Module != "" {
		rpcReq.Module = proto.String(req.Module)
	}
	if req.Operation != "" {
		rpcReq.Operation = proto.String(req.Operation)
	}
	if req.OperatorName != "" {
		rpcReq.OperatorName = proto.String(req.OperatorName)
	}
	if req.Path != "" {
		rpcReq.Path = proto.String(req.Path)
	}

	resp, err := l.svcCtx.SysLogs.GetOperationLogList(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	list := make([]map[string]interface{}, 0, len(resp.List))
	for _, log := range resp.List {
		item := map[string]interface{}{
			"id":           log.GetIdStr(),
			"idStr":        log.GetIdStr(),
			"module":       log.GetModule(),
			"operation":    log.GetOperation(),
			"requestUrl":   log.GetRequestUrl(),
			"status":       log.GetStatus(),
			"duration":     log.GetDuration(),
			"operatorName": log.GetOperatorName(),
			"operatorIp":   log.GetOperatorIp(),
		}
		if log.CreatedAt != nil {
			item["createdAt"] = timex.FormatUnix(*log.CreatedAt)
		}
		list = append(list, item)
	}

	data := map[string]interface{}{
		"list":     list,
		"total":    resp.Total,
		"page":     req.Page,
		"pageSize": req.PageSize,
	}
	dataBytes, _ := json.Marshal(data)
	return &types.BaseResp{
		Code: int(resp.Code),
		Msg:  resp.Msg,
		Data: json.RawMessage(dataBytes),
	}, nil
}
