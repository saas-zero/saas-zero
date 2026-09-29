package logic

import (
	"context"
	"strconv"
	"time"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
)

type CreateTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTenantLogic {
	return &CreateTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateTenantLogic) CreateTenant(req *types.TenantReq) (*types.BaseResp, error) {
	rpcReq := &apps.TenantReq{
		Name:     proto.String(req.Name),
		Code:     proto.String(req.Code),
		Status:   proto.String(req.Status),
		Username: proto.String(req.Username),
		Password: proto.String(req.Password),
	}
	if aid := parseId(req.AdminId); aid > 0 {
		rpcReq.AdminId = proto.Int64(aid)
	}
	if pid := parseId(req.ParentId); pid > 0 {
		rpcReq.ParentId = proto.Int64(pid)
	}
	if pkgId := parseId(req.PackageId); pkgId > 0 {
		rpcReq.PackageId = proto.Int64(pkgId)
	}
	if req.ExpiredAt != "" {
		// 兼容毫秒时间戳与 yyyy-MM-dd 日期
		if v, err := strconv.ParseInt(req.ExpiredAt, 10, 64); err == nil {
			rpcReq.ExpiredAt = proto.Int64(v)
		} else if t, err := time.Parse("2006-01-02", req.ExpiredAt); err == nil {
			rpcReq.ExpiredAt = proto.Int64(t.UnixMilli())
		}
	}
	if req.Remark != "" {
		rpcReq.Remark = proto.String(req.Remark)
	}
	resp, err := l.svcCtx.SysTenants.CreateTenant(l.ctx, rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: resp.GetData()}, nil
}
