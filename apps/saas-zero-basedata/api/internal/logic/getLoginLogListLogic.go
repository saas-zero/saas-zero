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

type GetLoginLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLoginLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLoginLogListLogic {
	return &GetLoginLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetLoginLogListLogic) GetLoginLogList(req *types.LogPageReq) (*types.BaseResp, error) {
	rpcReq := &apps.LogPageReq{
		Page: int32(req.Page),
		Size: int32(req.PageSize),
	}
	if req.Username != "" {
		rpcReq.Username = proto.String(req.Username)
	}
	if req.Status != "" {
		rpcReq.Status = proto.String(req.Status)
	}
	if req.Ip != "" {
		rpcReq.Ip = proto.String(req.Ip)
	}

	resp, err := l.svcCtx.SysLogs.GetLoginLogList(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}

	list := make([]map[string]interface{}, 0, len(resp.List))
	for _, log := range resp.List {
		item := map[string]interface{}{
			"id":       log.GetIdStr(),
			"idStr":    log.GetIdStr(),
			"username": log.GetUsername(),
			"loginIp":  log.GetLoginIp(),
			"status":   log.GetStatus(),
			"msg":      log.GetMsg(),
		}
		if log.LoginAt != nil {
			item["loginAt"] = timex.FormatUnix(*log.LoginAt)
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
