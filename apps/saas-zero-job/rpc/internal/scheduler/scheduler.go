package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
	"github.com/saas-zero/saas-zero-common/pkg/redis"
	"github.com/saas-zero/saas-zero-job/ent"
	"github.com/saas-zero/saas-zero-job/ent/sysjob"
	"github.com/saas-zero/saas-zero-job/ent/sysjoblog"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// 任务默认时区
	defaultTimeZone = "Asia/Shanghai"
	// 幂等锁过期(秒)：防止异常场景锁残留，应远大于单次执行最长时间
	lockTTLSeconds = 3600
	// 每任务并发锁的 key 前缀
	concurrencyKeyPrefix = "job:concurrent:"
	// 幂等锁 key 前缀
	idempotencyKeyPrefix = "job:fire:"
)

// Options 调度器初始化参数
type Options struct {
	DB               *ent.Client
	Redis            *redis.Client
	Registry         *Registry
	SyncIntervalSec  int
	LogRetentionDays int
	CleanupHour      int
	Node             string
	// Disable cleanup 关闭日志清理（测试用）
	DisableCleanup bool
}

// JobEntry 运行态任务包装：DB 记录 + cron entry + 当前状态
type JobEntry struct {
	Job      *ent.SysJob
	EntryID  cron.EntryID
	Schedule cron.Schedule
	LockKey  string // 并发锁 key
}

// Scheduler 定时任务调度器
// 架构：每实例各跑一个 cron，触发时用 Redis SETNX 幂等锁去重，
// 只有抢到锁的实例真正执行 —— 支持多实例部署且无 leader 选举。
type Scheduler struct {
	db   *ent.Client
	rds  *redis.Client
	reg  *Registry
	cron *cron.Cron
	node string

	mu      sync.RWMutex
	entries map[int64]*JobEntry // jobID -> entry

	syncInterval  time.Duration
	retentionDays int
	cleanupHour   int
	stopCh        chan struct{}
	stopOnce      sync.Once
}

// NewScheduler 创建调度器（不自动启动，需调用 Start）
func NewScheduler(opts Options) *Scheduler {
	if opts.Registry == nil {
		opts.Registry = NewRegistry()
	}
	// 解析器：SecondOptional 兼容 5/6 位表达式，Descriptor 支持 @every/@daily 等
	parser := cron.NewParser(
		cron.SecondOptional |
			cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow |
			cron.Descriptor,
	)
	c := cron.New(cron.WithParser(parser))
	return &Scheduler{
		db:            opts.DB,
		rds:           opts.Redis,
		reg:           opts.Registry,
		cron:          c,
		node:          opts.Node,
		entries:       make(map[int64]*JobEntry),
		syncInterval:  time.Duration(opts.SyncIntervalSec) * time.Second,
		retentionDays: opts.LogRetentionDays,
		cleanupHour:   opts.CleanupHour,
		stopCh:        make(chan struct{}),
	}
}

// Register 注册任务处理器（供业务代码 init 调用）
func (s *Scheduler) Register(code string, fn HandlerFunc) error {
	return s.reg.Register(code, fn)
}

// Registry 返回注册表（供 RPC logic 校验 handler 是否存在）
func (s *Scheduler) Registry() *Registry {
	return s.reg
}

// SyncNow 立即全量同步一次 DB → 调度器（创建/更新/删除后主动调用，不等 30s 轮询）
func (s *Scheduler) SyncNow() {
	s.syncFromDB()
}

// Node 返回当前实例标识
func (s *Scheduler) Node() string { return s.node }

// Start 启动调度器：
// 1. 加载所有 active 任务到 cron
// 2. 启动 30s 轮询协程（DB 对齐）
// 3. 启动日志清理协程（每日）
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.loadAllLocked(); err != nil {
		return fmt.Errorf("load jobs: %w", err)
	}
	s.cron.Start()

	go s.syncLoop()
	go s.cleanupLoop()
	logx.Infof("scheduler started, %d jobs loaded, sync interval=%s", len(s.entries), s.syncInterval)
	return nil
}

// Stop 停止调度器并清理
func (s *Scheduler) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
		if s.cron != nil {
			ctx := s.cron.Stop()
			select {
			case <-ctx.Done():
			case <-time.After(5 * time.Second):
			}
		}
	})
}

// ------------------------------------------------------------
// 加载与同步
// ------------------------------------------------------------

// loadAllLocked 全量加载 active 任务（调用方需持有 s.mu）
func (s *Scheduler) loadAllLocked() error {
	ctx := context.Background()
	jobs, err := s.db.SysJob.Query().
		Where(sysjob.StatusEQ(sysjob.StatusActive)).
		All(ctx)
	if err != nil {
		return err
	}
	loaded := make(map[int64]bool)
	for _, j := range jobs {
		if err := s.addEntryLocked(j); err != nil {
			logx.Errorf("load job %d(%s): %v", j.ID, j.Name, err)
			continue
		}
		loaded[j.ID] = true
	}
	// 移除已被删除/暂停的任务
	for id := range s.entries {
		if !loaded[id] {
			s.removeEntryLocked(id)
		}
	}
	return nil
}

// syncLoop 定期从 DB 拉取 active 任务做增量对齐
func (s *Scheduler) syncLoop() {
	if s.syncInterval <= 0 {
		s.syncInterval = 30 * time.Second
	}
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.syncFromDB()
		}
	}
}

func (s *Scheduler) syncFromDB() {
	ctx := context.Background()
	jobs, err := s.db.SysJob.Query().
		Where(sysjob.StatusEQ(sysjob.StatusActive)).
		All(ctx)
	if err != nil {
		logx.Errorf("sync jobs from db: %v", err)
		return
	}
	dbIDs := make(map[int64]bool, len(jobs))
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, j := range jobs {
		dbIDs[j.ID] = true
		if _, ok := s.entries[j.ID]; ok {
			// 已加载：检查是否更新（name/handler/cron 等变化需重建 entry）
			old := s.entries[j.ID].Job
			if old.Handler != j.Handler || old.CronExpression != j.CronExpression ||
				old.TimeZone != j.TimeZone || old.Concurrent != j.Concurrent {
				s.removeEntryLocked(j.ID)
				if err := s.addEntryLocked(j); err != nil {
					logx.Errorf("reload job %d(%s): %v", j.ID, j.Name, err)
				}
			}
		} else {
			if err := s.addEntryLocked(j); err != nil {
				logx.Errorf("add job %d(%s): %v", j.ID, j.Name, err)
			}
		}
	}
	// 删除 DB 中已不存在的 active 任务
	for id := range s.entries {
		if !dbIDs[id] {
			s.removeEntryLocked(id)
		}
	}
}

// ------------------------------------------------------------
// entry 管理（调用方需持有 s.mu）
// ------------------------------------------------------------

func (s *Scheduler) addEntryLocked(j *ent.SysJob) error {
	// per-job 时区：非默认时区用 CRON_TZ= 前缀，让 parser 构造带时区的 SpecSchedule；
	// cron 内部用绝对时间差等待，跨时区不会错。
	expr := j.CronExpression
	if j.TimeZone != "" && j.TimeZone != defaultTimeZone &&
		!strings.HasPrefix(expr, "CRON_TZ=") && !strings.HasPrefix(expr, "TZ=") {
		expr = "CRON_TZ=" + j.TimeZone + " " + expr
	}

	entryID, err := s.cron.AddJob(expr, cronJob(func() {
		s.fire(j.ID, triggerTypeCron)
	}))
	if err != nil {
		return fmt.Errorf("invalid cron %q: %w", j.CronExpression, err)
	}

	lockKey := concurrencyKeyPrefix + fmt.Sprintf("%d", j.ID)
	s.entries[j.ID] = &JobEntry{
		Job:     j,
		EntryID: entryID,
		LockKey: lockKey,
	}
	return nil
}

func (s *Scheduler) removeEntryLocked(jobID int64) {
	e, ok := s.entries[jobID]
	if !ok {
		return
	}
	s.cron.Remove(e.EntryID)
	delete(s.entries, jobID)
}

// ------------------------------------------------------------
// 触发与执行
// ------------------------------------------------------------

type triggerType string

const (
	triggerTypeCron   triggerType = "cron"
	triggerTypeManual triggerType = "manual"
)

// fire 任务触发入口（cron 回调）：misfire 检查 + 幂等锁 + 并发锁 + 执行
func (s *Scheduler) fire(jobID int64, tt triggerType) {
	ctx := context.Background()

	// cron 触发时检查 misfire_policy：next_run_at 已过期说明之前漏跑了
	if tt == triggerTypeCron {
		j, err := s.db.SysJob.Get(ctx, jobID)
		if err != nil {
			logx.Errorf("job %d misfire check load: %v", jobID, err)
			return
		}
		if !j.NextRunAt.IsZero() && j.NextRunAt.Before(time.Now()) {
			switch j.MisfirePolicy {
			case sysjob.MisfirePolicySkip:
				s.recordLog(ctx, j, tt, sysjoblog.StatusSkipped, 0, "missed and skip policy", "", 0)
				s.updateNextRunOnly(ctx, j)
				logx.Infof("job %d skipped (misfire_policy=skip, next_run_at=%s)", jobID, j.NextRunAt.Format(time.RFC3339))
				return
			case sysjob.MisfirePolicyFireAll:
				missed := countMissedRuns(j.NextRunAt, time.Now(), j.CronExpression)
				if missed > 1 {
					logx.Infof("job %d fire_all: %d missed runs", jobID, missed-1)
					for i := 0; i < missed-1; i++ {
						s.execute(ctx, jobID, tt)
					}
				}
				// 最后一次由下面的正常流程执行
			}
		}
	}

	// 幂等锁：同一调度轮次只允许一个实例执行（多实例去重）
	nextRun := time.Now().Unix()
	lockKey := fmt.Sprintf("%s%d:%d", idempotencyKeyPrefix, jobID, nextRun)
	ok, err := s.rds.SetNX(lockKey, s.node, lockTTLSeconds)
	if err != nil {
		logx.Errorf("job %d idempotency lock: %v", jobID, err)
		return
	}
	if !ok {
		logx.Infof("job %d already fired by another instance, skip", jobID)
		return
	}
	defer s.releaseLock(lockKey)

	s.execute(ctx, jobID, tt)
}

// execute 实际执行：加载任务 -> 并发检查 -> 执行 -> 写日志 -> 更新 last_*
func (s *Scheduler) execute(ctx context.Context, jobID int64, tt triggerType) {
	// 从 DB 重新加载，避免使用已过期快照
	j, err := s.db.SysJob.Get(ctx, jobID)
	if err != nil {
		logx.Errorf("job %d load: %v", jobID, err)
		return
	}

	// 手动触发不受状态限制；cron 触发仅 active 任务
	if tt == triggerTypeCron && j.Status != sysjob.StatusActive {
		return
	}

	s.mu.RLock()
	entry := s.entries[jobID]
	s.mu.RUnlock()

	// 并发检查：不允许并发时，同一任务已在执行则跳过
	if !j.Concurrent && entry != nil {
		lockKey := entry.LockKey
		ok, err := s.rds.SetNX(lockKey, s.node, lockTTLSeconds)
		if err != nil {
			logx.Errorf("job %d concurrency lock: %v", jobID, err)
			s.recordLog(ctx, j, tt, sysjoblog.StatusFail, 0, "", fmt.Sprintf("concurrency lock error: %v", err), 0)
			return
		}
		if !ok {
			// 已在执行，跳过本轮
			s.recordLog(ctx, j, tt, sysjoblog.StatusSkipped, 0, "previous run still in progress", "", 0)
			return
		}
		defer s.releaseLock(lockKey)
	}

	start := time.Now()
	attempt, runErr := s.runWithRetry(ctx, j, tt)
	duration := time.Since(start).Milliseconds()
	s.updateJobAfterRun(ctx, j, runErr, duration)
	s.recordFinalLog(ctx, j, tt, attempt, runErr, duration)
}

// runWithRetry 带重试地执行 handler，返回 (最终尝试次数, 最后错误)。
// 日志统一由 execute/recordFinalLog 写一条终态（不在此处记日志），
// 由 execute 携带真实 duration。
func (s *Scheduler) runWithRetry(ctx context.Context, j *ent.SysJob, tt triggerType) (int32, error) {
	fn, ok := s.reg.Lookup(j.Handler)
	if !ok {
		return 1, fmt.Errorf("handler %q not registered", j.Handler)
	}

	params := map[string]any{}
	if j.Params != "" {
		if err := json.Unmarshal([]byte(j.Params), &params); err != nil {
			return 1, fmt.Errorf("invalid params json: %w", err)
		}
	}

	maxAttempts := j.MaxRetry + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	var lastErr error
	var lastAttempt int32
	for attempt := int32(1); attempt <= maxAttempts; attempt++ {
		lastAttempt = attempt
		// 每次尝试带超时 context
		timeout := time.Duration(j.Timeout) * time.Second
		if timeout <= 0 {
			timeout = 30 * time.Second
		}
		execCtx, cancel := context.WithTimeout(contextWithSystemUser(ctx, j), timeout)
		err := fn(execCtx, params)
		cancel()

		if err == nil {
			return lastAttempt, nil
		}
		lastErr = err
		if attempt >= maxAttempts {
			break
		}
		// 等待重试间隔
		interval := time.Duration(j.RetryInterval) * time.Second
		if interval <= 0 {
			interval = 10 * time.Second
		}
		select {
		case <-ctx.Done():
			return lastAttempt, lastErr
		case <-time.After(interval):
		}
	}
	return lastAttempt, lastErr
}

// recordFinalLog 写一条终态日志（success/fail/timeout），携带真实执行时长
func (s *Scheduler) recordFinalLog(ctx context.Context, j *ent.SysJob, tt triggerType, attempt int32, runErr error, duration int64) {
	status := sysjoblog.StatusSuccess
	var exception string
	if runErr != nil {
		if errors.Is(runErr, context.DeadlineExceeded) {
			status = sysjoblog.StatusTimeout
		} else {
			status = sysjoblog.StatusFail
		}
		exception = runErr.Error()
	}
	s.recordLog(ctx, j, tt, status, attempt, "", exception, duration)
}

// contextWithSystemUser 注入 system 用户上下文（审计字段填充）
// job 为平台级（无租户），调度器写日志/更新 last_* 时必须显式设置用户
func contextWithSystemUser(ctx context.Context, j *ent.SysJob) context.Context {
	// TenantMixin 已从 job 表移除，只需用户，不需要租户
	ctx = mixins.SetCurrentUserId(ctx, 1)
	ctx = mixins.SetCurrentUserName(ctx, "system")
	return ctx
}

func (s *Scheduler) releaseLock(key string) {
	if _, err := s.rds.Del(key); err != nil {
		logx.Errorf("release lock %s: %v", key, err)
	}
}

// updateJobAfterRun 更新 last_* 与 next_run_at（成功/失败两种终态）
func (s *Scheduler) updateJobAfterRun(ctx context.Context, j *ent.SysJob, runErr error, duration int64) {
	ctx = contextWithSystemUser(ctx, j)
	upd := s.db.SysJob.UpdateOneID(j.ID).
		SetLastRunAt(time.Now()).
		SetLastDuration(duration)
	if runErr != nil {
		upd = upd.SetLastStatus(string(sysjoblog.StatusFail))
		msg := runErr.Error()
		if len(msg) > 500 {
			msg = msg[:500]
		}
		upd = upd.SetLastError(msg)
	} else {
		upd = upd.SetLastStatus(string(sysjoblog.StatusSuccess)).ClearLastError()
	}
	// 回写下次触发时间：从调度条目计算（手动触发无条目则跳过）
	s.mu.RLock()
	entry := s.entries[j.ID]
	s.mu.RUnlock()
	if entry != nil && entry.Schedule != nil {
		if next := entry.Schedule.Next(time.Now()); !next.IsZero() {
			upd = upd.SetNextRunAt(next)
		}
	}
	if _, err := upd.Save(ctx); err != nil {
		logx.Errorf("update job %d last_*: %v", j.ID, err)
	}
}

// recordLog 写一条执行日志
func (s *Scheduler) recordLog(ctx context.Context, j *ent.SysJob, tt triggerType, status sysjoblog.Status, attempt int32, msg, exception string, duration int64) {
	ctx = contextWithSystemUser(ctx, j)
	if len(msg) > 500 {
		msg = msg[:500]
	}
	if len(exception) > 1000 {
		exception = exception[:1000]
	}
	_, err := s.db.SysJobLog.Create().
		SetJobID(j.ID).
		SetJobName(j.Name).
		SetJobGroup(j.Group).
		SetHandler(j.Handler).
		SetTriggerType(sysjoblog.TriggerType(tt)).
		SetExecNode(s.node).
		SetStatus(status).
		SetAttempt(attempt).
		SetMessage(msg).
		SetExceptionInfo(exception).
		SetDuration(duration).
		Save(ctx)
	if err != nil {
		logx.Errorf("record job log: %v", err)
	}
}

// RunNow 手动触发执行（RPC RunJobOnce）
// 不受任务状态限制，记录 trigger=manual；
// 若任务不允许并发且正在运行，则返回错误
func (s *Scheduler) RunNow(ctx context.Context, jobID int64) error {
	j, err := s.db.SysJob.Get(ctx, jobID)
	if err != nil {
		return err
	}
	// 手动触发不加幂等锁（用户主动触发），但并发锁仍生效
	s.mu.RLock()
	entry := s.entries[jobID]
	s.mu.RUnlock()

	// 并发锁：不允许并发时，同一任务已在执行则跳过（RunNow 手动触发不检查：用户显式操作）
	if !j.Concurrent {
		if entry != nil {
			ok, err := s.rds.SetNX(entry.LockKey, s.node, lockTTLSeconds)
			if err != nil {
				return fmt.Errorf("concurrency lock: %w", err)
			}
			if !ok {
				return fmt.Errorf("job is already running")
			}
			defer s.releaseLock(entry.LockKey)
		}
	}

	start := time.Now()
	attempt, runErr := s.runWithRetry(ctx, j, triggerTypeManual)
	duration := time.Since(start).Milliseconds()
	s.updateJobAfterRun(ctx, j, runErr, duration)
	s.recordFinalLog(ctx, j, triggerTypeManual, attempt, runErr, duration)
	return runErr
}

// CleanLogs 清理超期日志（按保留天数），供日志清理协程与 RPC 调用
func (s *Scheduler) CleanLogs(keepDays int) (int, error) {
	if keepDays <= 0 {
		return 0, fmt.Errorf("invalid keep days: %d", keepDays)
	}
	cutoff := time.Now().AddDate(0, 0, -keepDays)
	deleted, err := s.db.SysJobLog.Delete().
		Where(sysjoblog.CreatedAtLT(cutoff)).
		Exec(context.Background())
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

// cleanupLoop 每日清理过期日志
func (s *Scheduler) cleanupLoop() {
	if s.retentionDays <= 0 {
		s.retentionDays = 30
	}
	// 计算下次清理时间
	next := s.nextCleanupTime()
	timer := time.NewTimer(time.Until(next))
	defer timer.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-timer.C:
			n, err := s.CleanLogs(s.retentionDays)
			if err != nil {
				logx.Errorf("clean logs: %v", err)
			} else {
				logx.Infof("cleaned %d expired job logs (retention=%dd)", n, s.retentionDays)
			}
			// 重设下次
			timer.Reset(time.Until(s.nextCleanupTime()))
		}
	}
}

func (s *Scheduler) nextCleanupTime() time.Time {
	now := time.Now()
	hour := s.cleanupHour
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	if !next.After(now) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

// FireNow 内部触发（测试/开发用），模拟一次 cron 触发
func (s *Scheduler) FireNow(ctx context.Context, jobID int64) error {
	s.fire(jobID, triggerTypeManual)
	return nil
}

// updateNextRunOnly 只更新 next_run_at，不执行（misfire skip 用）
func (s *Scheduler) updateNextRunOnly(ctx context.Context, j *ent.SysJob) {
	ctx = contextWithSystemUser(ctx, j)
	s.mu.RLock()
	entry := s.entries[j.ID]
	s.mu.RUnlock()
	if entry != nil && entry.Schedule != nil {
		if next := entry.Schedule.Next(time.Now()); !next.IsZero() {
			if _, err := s.db.SysJob.UpdateOneID(j.ID).SetNextRunAt(next).Save(ctx); err != nil {
				logx.Errorf("job %d update next_run_at: %v", j.ID, err)
			}
			return
		}
	}
	// 无 entry 时，用 cron 解析器算下次时间
	parser := cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)
	schedule, err := parser.Parse(j.CronExpression)
	if err != nil {
		logx.Errorf("job %d parse cron %q for next_run: %v", j.ID, j.CronExpression, err)
		return
	}
	next := schedule.Next(time.Now())
	if !next.IsZero() {
		if _, err := s.db.SysJob.UpdateOneID(j.ID).SetNextRunAt(next).Save(ctx); err != nil {
			logx.Errorf("job %d update next_run_at: %v", j.ID, err)
		}
	}
}

// countMissedRuns 计算从 from 到 to 之间 cron 表达式触发了多少次
func countMissedRuns(from, to time.Time, cronExpr string) int {
	if !from.Before(to) {
		return 0
	}
	parser := cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)
	schedule, err := parser.Parse(cronExpr)
	if err != nil {
		return 0
	}
	count := 0
	t := from
	for t.Before(to) {
		t = schedule.Next(t)
		if t.Before(to) || t.Equal(to) {
			count++
		}
	}
	return count
}

// helper：cronJob 适配 cron.FuncJob
type cronJob func()

func (f cronJob) Run() { f() }
