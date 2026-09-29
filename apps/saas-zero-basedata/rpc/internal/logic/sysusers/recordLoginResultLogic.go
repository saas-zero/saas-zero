package sysuserslogic

import (
	"context"
	"time"

	"github.com/saas-zero/saas-zero-basedata/ent/sysloginlog"
	"github.com/saas-zero/saas-zero-basedata/ent/sysuser"
	"github.com/saas-zero/saas-zero-basedata/rpc/apps"
	"github.com/saas-zero/saas-zero-basedata/rpc/internal/svc"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/zeromicro/go-zero/core/logx"
)

// 登录失败锁定策略：连续失败达到 maxLoginErrors 次后锁定 lockoutDuration。
const (
	maxLoginErrors  = 5
	lockoutDuration = 30 * time.Minute
)

type RecordLoginResultLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecordLoginResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordLoginResultLogic {
	return &RecordLoginResultLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// RecordLoginResult 在一次登录尝试后被调用：既更新用户锁定计数和最近登录信息，
// 又写入一条登录审计日志（成功或失败）。
func (l *RecordLoginResultLogic) RecordLoginResult(in *apps.LoginRecordReq) (*apps.EmptyResp, error) {
	now := time.Now()
	userId := in.GetUserId()
	tenantId := in.GetTenantId()

	if userId > 0 {
		target, err := l.svcCtx.DB.SysUser.TenantQuery(tenantId).
			Where(sysuser.IDEQ(userId)).
			Only(l.ctx)
		if err != nil {
			return nil, err
		}
		update := l.svcCtx.DB.SysUser.UpdateOne(target)
		if in.GetSuccess() {
			// 成功：清零错误计数、解除锁定、记录最近登录 IP/时间
			update.SetLoginErrorCount(0).
				ClearLockoutUntil().
				SetLoginIP(in.GetIp()).
				SetLoginAt(now)
		} else {
			// 失败：读取当前错误次数，累加后判断是否需要锁定
			newCount := target.LoginErrorCount + 1
			update.SetLoginErrorCount(newCount)
			if newCount >= maxLoginErrors {
				update.SetLockoutUntil(now.Add(lockoutDuration))
			}
		}
		if err := update.Exec(l.ctx); err != nil {
			return nil, err
		}
	}

	// 写登录日志（sys_login_logs 无 TenantMixin，自建 tenant_id 业务字段，需显式设置）
	status := sysloginlog.StatusSuccess
	if !in.GetSuccess() {
		status = sysloginlog.StatusFail
	}
	logUserId := userId
	if logUserId <= 0 {
		logUserId = 1 // 用户不存在时占位，满足 user_id Positive 约束
	}
	if _, err := l.svcCtx.DB.SysLoginLog.Create().
		SetUserID(logUserId).
		SetUsername(in.GetUsername()).
		SetIP(in.GetIp()).
		SetStatus(status).
		SetMessage(in.GetMessage()).
		SetLoginTime(now).
		SetTenantID(tenantId).
		Save(l.ctx); err != nil {
		// 日志写入失败不应阻断登录流程，仅记录错误
		logx.WithContext(l.ctx).Errorf("record login log error: %v", err)
	}

	return &apps.EmptyResp{Code: int32(errno.Success.Code), Msg: errno.Success.Msg}, nil
}
