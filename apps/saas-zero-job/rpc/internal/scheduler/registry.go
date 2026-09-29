package scheduler

import (
	"context"
	"fmt"
	"sync"
)

// HandlerFunc 定时任务处理器。
// params 来自 sys_jobs.params（JSON 解析后的 map），
// ctx 已带有 job 的超时（job.Timeout 秒）与 system 用户上下文（审计字段填充）。
type HandlerFunc func(ctx context.Context, params map[string]any) error

// Registry 进程内 handler 注册表。
// 业务代码在 init() 中调用 job.Register("report.daily", fn) 注册；
// 调度器通过 DB 中 job.handler 的值查表执行。
// 未注册的 code 执行时记为失败（handler not found），不 panic。
type Registry struct {
	mu sync.RWMutex
	m  map[string]HandlerFunc
}

func NewRegistry() *Registry {
	return &Registry{m: make(map[string]HandlerFunc)}
}

// Register 注册处理器，重复注册返回错误（防止同名误覆盖）。
func (r *Registry) Register(code string, fn HandlerFunc) error {
	if code == "" || fn == nil {
		return fmt.Errorf("invalid handler registration: code=%q fn=nil", code)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.m[code]; dup {
		return fmt.Errorf("job handler %q already registered", code)
	}
	r.m[code] = fn
	return nil
}

// Lookup 根据注册码查找处理器。
func (r *Registry) Lookup(code string) (HandlerFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, ok := r.m[code]
	return fn, ok
}

// Has 判断是否已注册。
func (r *Registry) Has(code string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.m[code]
	return ok
}

// Keys 返回所有已注册的 handler code 列表。
func (r *Registry) Keys() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys := make([]string, 0, len(r.m))
	for k := range r.m {
		keys = append(keys, k)
	}
	return keys
}
