package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/errno"
	"github.com/saas-zero/saas-zero-job/ent"
	"github.com/saas-zero/saas-zero-job/rpc/apps"
	"github.com/saas-zero/saas-zero-job/rpc/internal/scheduler"
)

// errInvalidParam 业务参数错误
func errInvalidParam(format string, args ...any) error {
	return errno.New(errno.InvalidParam.Code, fmt.Sprintf(format, args...))
}

// validateCron 校验 cron 表达式（兼容 5/6 位：秒可选 + 描述符）
func validateCron(expr string) error {
	parser := cron.NewParser(cron.SecondOptional |
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow |
		cron.Descriptor)
	_, err := parser.Parse(expr)
	return err
}

// validateParams 校验 params 为合法 JSON 对象
func validateParams(params string) error {
	if params == "" || params == "{}" {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(params), &m); err != nil {
		return errInvalidParam("params must be a valid JSON object: %v", err)
	}
	return nil
}

// parseID string → int64（前端精度：proto 全 string）
func parseID(s string) (int64, error) {
	if s == "" {
		return 0, errInvalidParam("id is required")
	}
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, errInvalidParam("invalid id: %s", s)
	}
	return id, nil
}

// jobToProto ent.SysJob → apps.Job
func jobToProto(j *ent.SysJob) *apps.Job {
	p := &apps.Job{
		Id:             strconv.FormatInt(j.ID, 10),
		Name:           j.Name,
		Group:          j.Group,
		Handler:        j.Handler,
		Params:         j.Params,
		CronExpression: j.CronExpression,
		TimeZone:       j.TimeZone,
		Concurrent:     j.Concurrent,
		Timeout:        j.Timeout,
		MaxRetry:       j.MaxRetry,
		RetryInterval:  j.RetryInterval,
		Status:         string(j.Status),
		Remark:         j.Remark,
		LastStatus:     j.LastStatus,
		LastDuration:   j.LastDuration,
		LastError:      j.LastError,
	}
	if !j.CreatedAt.IsZero() {
		p.CreatedAt = j.CreatedAt.Unix()
	}
	if !j.UpdatedAt.IsZero() {
		p.UpdatedAt = j.UpdatedAt.Unix()
	}
	if !j.NextRunAt.IsZero() {
		p.NextRunAt = j.NextRunAt.Unix()
	}
	if !j.LastRunAt.IsZero() {
		p.LastRunAt = j.LastRunAt.Unix()
	}
	if j.MisfirePolicy != "" {
		p.MisfirePolicy = string(j.MisfirePolicy)
	} else {
		p.MisfirePolicy = "fire_once" // 默认值
	}
	return p
}

// logToProto ent.SysJobLog → apps.JobLog
func logToProto(l *ent.SysJobLog) *apps.JobLog {
	p := &apps.JobLog{
		Id:            strconv.FormatInt(l.ID, 10),
		JobId:         strconv.FormatInt(l.JobID, 10),
		JobName:       l.JobName,
		JobGroup:      l.JobGroup,
		Handler:       l.Handler,
		TriggerType:   string(l.TriggerType),
		ExecNode:      l.ExecNode,
		Status:        string(l.Status),
		Attempt:       l.Attempt,
		Message:       l.Message,
		ExceptionInfo: l.ExceptionInfo,
		Duration:      l.Duration,
	}
	if !l.CreatedAt.IsZero() {
		p.CreatedAt = l.CreatedAt.Unix()
	}
	return p
}

// systemCtx 注入 system 用户上下文（无租户：job 为平台级）
// RPC 侧校验层需要审计用户时使用；业务侧从 gRPC metadata 注入的 ctx 已有用户。
func systemCtx(ctx context.Context) context.Context {
	if mixins.GetCurrentUserId(ctx) > 0 {
		return ctx
	}
	ctx = mixins.SetCurrentUserId(ctx, 1)
	return mixins.SetCurrentUserName(ctx, "system")
}

// handlerRegistered 检查处理器是否已注册
func handlerRegistered(reg *scheduler.Registry, code string) bool {
	if reg == nil || code == "" {
		return false
	}
	return reg.Has(code)
}

// normalizePage 分页参数标准化：page>=1, size 1..100，默认 1/20
func normalizePage(page, size int32) (int, int) {
	p := int(page)
	if p < 1 {
		p = 1
	}
	s := int(size)
	if s < 1 {
		s = 20
	}
	if s > 100 {
		s = 100
	}
	return p, s
}

// unixTime 秒级时间戳 → time.Time
func unixTime(sec int64) time.Time {
	if sec <= 0 {
		return time.Time{}
	}
	return time.Unix(sec, 0)
}
