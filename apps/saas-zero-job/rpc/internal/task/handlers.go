package task

import (
	"context"
	"fmt"
	"time"

	"github.com/saas-zero/saas-zero-job/rpc/internal/scheduler"
	"github.com/zeromicro/go-zero/core/logx"
)

// RegisterAll 注册所有内置任务处理器。
// 业务代码在此处通过 job.Register(code, fn) 注册，code 与 sys_jobs.handler 字段对应。
func RegisterAll(reg *scheduler.Registry) error {
	// 示例任务：hello 打印参数（演示注册表机制，验证调度器工作）
	if err := reg.Register("demo.hello", func(ctx context.Context, params map[string]any) error {
		name := "world"
		if v, ok := params["name"].(string); ok && v != "" {
			name = v
		}
		logx.Infof("demo.hello executed, name=%s", name)
		return nil
	}); err != nil {
		return err
	}

	// 示例任务：delay 模拟耗时操作（配合 timeout/concurrent 验证超时与并发锁）
	if err := reg.Register("demo.delay", func(ctx context.Context, params map[string]any) error {
		seconds := 3
		if v, ok := params["seconds"].(float64); ok && v > 0 {
			seconds = int(v)
		}
		select {
		case <-time.After(time.Duration(seconds) * time.Second):
			return nil
		case <-ctx.Done():
			return fmt.Errorf("cancelled by timeout: %w", ctx.Err())
		}
	}); err != nil {
		return err
	}

	// 示例任务：fail 总是失败（配合 max_retry 验证重试）
	if err := reg.Register("demo.fail", func(ctx context.Context, params map[string]any) error {
		return fmt.Errorf("demo.fail always fails (params=%v)", params)
	}); err != nil {
		return err
	}

	return nil
}
