package scheduler

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// HandlerFunc 定时任务处理器。
// params 来自 sys_jobs.params（JSON 解析后的 map），
// ctx 已带有 job 的超时（job.Timeout 秒）与 system 用户上下文（审计字段填充）。
type HandlerFunc func(ctx context.Context, params map[string]any) error

// Item 处理器注册信息（不含执行函数，供管理端展示）
type Item struct {
	Code string
	Name string
}

// Registry 进程内 handler 注册表。
// 业务代码在 init() 中调用 job.Register(code, name, fn) 注册；
// 调度器通过 DB 中 job.handler 的值查表执行，
// 管理端（/system/job/handlers）通过 Handlers() 取可选列表——唯一来源，无配置副本。
// 未注册的 code 执行时记为失败（handler not found），不 panic。
type Registry struct {
	mu sync.RWMutex
	m  map[string]entry
}

// entry 注册项：展示信息 + 执行函数
type entry struct {
	item Item
	fn   HandlerFunc
}

func NewRegistry() *Registry {
	return &Registry{m: make(map[string]entry)}
}

// Register 注册处理器，重复注册返回错误（防止同名误覆盖）。
// name 为管理端展示名，留空时退化为 code。
func (r *Registry) Register(code, name string, fn HandlerFunc) error {
	if code == "" || fn == nil {
		return fmt.Errorf("invalid handler registration: code=%q fn=nil", code)
	}
	if name == "" {
		name = code
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.m[code]; dup {
		return fmt.Errorf("job handler %q already registered", code)
	}
	r.m[code] = entry{item: Item{Code: code, Name: name}, fn: fn}
	return nil
}

// Lookup 根据注册码查找处理器。
func (r *Registry) Lookup(code string) (HandlerFunc, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.m[code]
	return e.fn, ok
}

// Has 判断是否已注册。
func (r *Registry) Has(code string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.m[code]
	return ok
}

// Handlers 返回全部已注册处理器（按 code 排序，便于展示与测试比对）。
func (r *Registry) Handlers() []Item {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]Item, 0, len(r.m))
	for _, e := range r.m {
		items = append(items, e.item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Code < items[j].Code })
	return items
}
