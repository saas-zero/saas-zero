package logic

import (
	"context"
	"strconv"
	"time"

	"github.com/saas-zero/saas-zero-basedata/api/internal/svc"
	"github.com/saas-zero/saas-zero-basedata/api/internal/types"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type InitLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitLogic {
	return &InitLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InitLogic) initCtx() context.Context {
	return metadata.NewOutgoingContext(l.ctx, metadata.Pairs(
		"x-user-id", "1",
		"x-user-name", "system",
		"x-tenant-id", "1",
	))
}

func (l *InitLogic) InitCreatePackage(req *types.PackageReq) (*types.BaseResp, error) {
	rpcReq := &apps.PackageReq{
		Name:   proto.String(req.Name),
		Code:   proto.String(req.Code),
		Status: proto.String(req.Status),
	}
	if req.Sort > 0 {
		rpcReq.Sort = proto.Int32(req.Sort)
	}
	if req.Remark != "" {
		rpcReq.Remark = proto.String(req.Remark)
	}
	resp, err := l.svcCtx.SysPackages.CreatePackage(l.initCtx(), rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: resp.GetData()}, nil
}

func (l *InitLogic) InitCreateTenant(req *types.TenantReq) (*types.BaseResp, error) {
	rpcReq := &apps.TenantReq{
		Name:   proto.String(req.Name),
		Code:   proto.String(req.Code),
		Status: proto.String(req.Status),
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
	resp, err := l.svcCtx.SysTenants.CreateTenant(l.initCtx(), rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: resp.GetData()}, nil
}

func (l *InitLogic) InitCreateUser(req *types.UserReq) (*types.BaseResp, error) {
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
	resp, err := l.svcCtx.SysUsers.CreateUser(l.initCtx(), rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: resp.GetData()}, nil
}

func (l *InitLogic) InitAll() (*types.BaseResp, error) {
	resp, err := l.svcCtx.SysInit.InitAll(l.initCtx(), &apps.EmptyReq{})
	if err != nil {
		return nil, err
	}
	l.reloadCasbin()
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg}, nil
}

// reloadCasbin forces the API Casbin enforcer to reload policies from DB.
// Must be called after any RPC operation that modifies Casbin policies.
func (l *InitLogic) reloadCasbin() {
	if l.svcCtx.Enforcer != nil {
		if err := l.svcCtx.Enforcer.LoadPolicy(); err != nil {
			logx.Errorf("reloadCasbin: failed to reload casbin policies: %v", err)
		}
	}
}

func (l *InitLogic) InitCreateRole(req *types.RoleReq) (*types.BaseResp, error) {
	rpcReq := &apps.RoleReq{
		Name:   proto.String(req.Name),
		Code:   proto.String(req.Code),
		Status: proto.String(req.Status),
	}
	if req.Sort > 0 {
		rpcReq.Sort = proto.Int32(req.Sort)
	}
	if req.Remark != "" {
		rpcReq.Remark = proto.String(req.Remark)
	}
	if len(req.MenuIds) > 0 {
		rpcReq.MenuIds = parseIds(req.MenuIds)
	}
	if len(req.ApiIds) > 0 {
		rpcReq.ApiIds = parseIds(req.ApiIds)
	}
	resp, err := l.svcCtx.SysRoles.CreateRole(l.initCtx(), rpcReq)
	if err != nil {
		return nil, err
	}
	return &types.BaseResp{Code: int(resp.Code), Msg: resp.Msg, Data: resp.GetData()}, nil
}
