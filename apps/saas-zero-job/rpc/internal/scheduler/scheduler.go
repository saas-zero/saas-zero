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
	// Disable cleanup 关闭日志清理（测试/排障用）
	DisableCleanup bool
}

// JobEntry 运行态任务包装：DB 记录 + cron entry + 当前状态
type JobEntry struct {
	Job      *ent.SysJob
	EntryID  cron.EntryID
	Schedule cron.Schedule
	// NextTick 下一个计划触发时刻（调度器自维护，多实例同一表达式得出同一序列，用作幂等锁粒度）
	NextTick time.Time
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
	// parser 与 cron 实例共用同一解析器（5/6 位 + 描述符 + 时区前缀）
	parser cron.Parser

	mu      sync.RWMutex
	entries map[int64]*JobEntry // jobID -> entry

	syncInterval   time.Duration
	retentionDays  int
	cleanupHour    int
	disableCleanup bool
	stopCh         chan struct{}
	stopOnce       sync.Once
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
		db:             opts.DB,
		rds:            opts.Redis,
		reg:            opts.Registry,
		cron:           c,
		node:           opts.Node,
		parser:         parser,
		entries:        make(map[int64]*JobEntry),
		syncInterval:   time.Duration(opts.SyncIntervalSec) * time.Second,
		retentionDays:  opts.LogRetentionDays,
		cleanupHour:    opts.CleanupHour,
		disableCleanup: opts.DisableCleanup,
		stopCh:         make(chan struct{}),
	}
}

// Register 注册任务处理器（供业务代码 init 调用）
func (s *Scheduler) Register(code, name string, fn HandlerFunc) error {
	return s.reg.Register(code, name, fn)
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
	if s.disableCleanup {
		logx.Info("job log cleanup disabled by config")
	} else {
		go s.cleanupLoop()
	}
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
	sched, err := s.parseSchedule(j)
	if err != nil {
		return err
	}

	entryID := s.cron.Schedule(sched, cronJob(func() {
		s.fire(j.ID, triggerTypeCron)
	}))

	lockKey := concurrencyKeyPrefix + fmt.Sprintf("%d", j.ID)
	s.entries[j.ID] = &JobEntry{
		Job:      j,
		EntryID:  entryID,
		Schedule: sched,
		NextTick: sched.Next(time.Now()),
		LockKey:  lockKey,
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

// fire 任务触发入口（cron 回调）：幂等锁 → 漏跑处理 → 执行
//
// 幂等锁以“计划触发时刻”为粒度（由调度器自维护的链路给出，不取本地时钟），
// 多实例即使本地时钟有偏差，同一调度轮次也命中同一把 key，从而只执行一次。
func (s *Scheduler) fire(jobID int64, tt triggerType) {
	ctx := context.Background()

	if tt == triggerTypeCron {
		// 本轮计划时刻 + 下一轮时刻（链路自建，多实例一致）
		tick, next := s.advanceNext(jobID)
		if tick.IsZero() {
			// entry 已不在调度表中（任务被暂停/删除，或 sync 已移除）：回调已过期，不再执行。
			// 若任务仍应调度，syncFromDB 会重新 addEntryLocked。
			logx.Errorf("job %d fired without schedule entry, dropped", jobID)
			return
		}

		lockKey := idempotencyKey(jobID, tick)
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

		// 只有抢到锁的实例处理漏跑并回写 next_run_at，避免多实例重复补跑
		j, err := s.db.SysJob.Get(ctx, jobID)
		if err != nil {
			logx.Errorf("job %d load for misfire: %v", jobID, err)
		} else {
			// 漏跑判定基准是上轮回写的 next_run_at，必须在覆盖它之前处理
			s.handleMissed(ctx, j, tick)
			// 回写下一轮计划时刻：本轮回写的是“下一个应触发时刻”，
			// 长任务执行期间该值不会滞后，因此漏跑判定不受执行时长影响。
			if !next.IsZero() {
				if _, err := s.db.SysJob.UpdateOneID(jobID).SetNextRunAt(next).
					Save(contextWithSystemUser(ctx, j)); err != nil {
					logx.Errorf("job %d write next_run_at: %v", jobID, err)
				}
			}
		}
	}

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

// updateJobAfterRun 更新 last_*（成功/失败两种终态）
// next_run_at 不在这里回写：它在 cron 触发时由 fire 写入“下一个计划时刻”，
// 与执行时长无关（长任务不会让 next_run_at 滞后而误判漏跑）。
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

// parseSchedule 解析任务 cron 表达式（兼容 5/6 位 + 描述符）
// per-job 时区：非默认时区加 CRON_TZ= 前缀，让 parser 构造带时区的 SpecSchedule；
// cron 内部用绝对时间差等待，跳时区不会错。
func (s *Scheduler) parseSchedule(j *ent.SysJob) (cron.Schedule, error) {
	expr := j.CronExpression
	if j.TimeZone != "" && j.TimeZone != defaultTimeZone &&
		!strings.HasPrefix(expr, "CRON_TZ=") && !strings.HasPrefix(expr, "TZ=") {
		expr = "CRON_TZ=" + j.TimeZone + " " + expr
	}
	sched, err := s.parser.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron %q: %w", j.CronExpression, err)
	}
	return sched, nil
}

// NextRunAt 用表达式推算任务的下次触发时刻（StartJob 恢复任务时重置漏跑窗口用）
func (s *Scheduler) NextRunAt(j *ent.SysJob) time.Time {
	if j == nil {
		return time.Time{}
	}
	sched, err := s.parseSchedule(j)
	if err != nil {
		logx.Errorf("job %d parse cron %q for next_run_at: %v", j.ID, j.CronExpression, err)
		return time.Time{}
	}
	return sched.Next(time.Now())
}

// advanceNext 取出本轮计划触发时刻并把链路推进到下一轮，返回 (本轮时刻, 下一轮时刻)。
// 链路由调度器自维护而不读 cron 内部 Entry.Prev/Next：run 循环先起任务 goroutine
// 再更新 Prev，从业务 goroutine 读有竞态；链路值由表达式递推，天然多实例一致。
func (s *Scheduler) advanceNext(jobID int64) (time.Time, time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[jobID]
	if !ok || e.Schedule == nil {
		return time.Time{}, time.Time{}
	}
	now := time.Now()
	tick := e.NextTick
	if tick.IsZero() {
		tick = now
	}
	e.NextTick = e.Schedule.Next(now)
	return tick, e.NextTick
}

// idempotencyKey 幂等锁 key：同一任务 + 同一计划时刻 → 同一 key（多实例去重）
func idempotencyKey(jobID int64, tick time.Time) string {
	return fmt.Sprintf("%s%d:%d", idempotencyKeyPrefix, jobID, tick.Unix())
}

// handleMissed 按 misfire_policy 处理漏跑轮次（当前轮次仍由 fire 正常执行）。
// 基准是 sys_jobs.next_run_at（上轮回写的“下一个计划时刻”）：只有“应触发却没触发”
// 的轮次才算漏跑（进程停机/实例阻塞）；暂停后恢复不算漏跑（StartJob 会重置基准）。
// 漏跑轮次计算见 missedRunsBefore（已扣除本轮自身）。
func (s *Scheduler) handleMissed(ctx context.Context, j *ent.SysJob, tick time.Time) {
	if j.NextRunAt.IsZero() {
		return
	}
	sched, err := s.parseSchedule(j)
	if err != nil {
		logx.Errorf("job %d misfire check parse cron: %v", j.ID, err)
		return
	}
	missed := missedRunsBefore(sched, j.NextRunAt, tick)
	if missed <= 0 {
		return
	}
	switch j.MisfirePolicy {
	case sysjob.MisfirePolicySkip:
		s.recordLog(ctx, j, triggerTypeCron, sysjoblog.StatusSkipped, 0,
			fmt.Sprintf("discard %d missed run(s) (misfire_policy=skip)", missed), "", 0)
		logx.Infof("job %d discarded %d missed run(s) (misfire_policy=skip)", j.ID, missed)
	case sysjob.MisfirePolicyFireOnce:
		logx.Infof("job %d catch up 1 of %d missed run(s) (misfire_policy=fire_once)", j.ID, missed)
		s.execute(ctx, j.ID, triggerTypeCron)
	case sysjob.MisfirePolicyFireAll:
		n := missed
		if n > maxCatchUpRuns {
			logx.Infof("job %d missed %d run(s), catch-up capped at %d (misfire_policy=fire_all)", j.ID, missed, maxCatchUpRuns)
			n = maxCatchUpRuns
		} else {
			logx.Infof("job %d catch up %d missed run(s) (misfire_policy=fire_all)", j.ID, n)
		}
		for i := 0; i < n; i++ {
			s.execute(ctx, j.ID, triggerTypeCron)
		}
	}
}

// maxCatchUpRuns 单轮补跑上限：防止长时间停摆后一次性补跑过多（fire_all）
const maxCatchUpRuns = 10

// missedRunsBefore 本轮之前的真正漏跑轮次（正常执行时为 0）：
// missedRuns 统计 (prevNextRunAt, tick] 内的计划时刻数，其中 tick 属于本轮正常执行，故减 1。
func missedRunsBefore(sched cron.Schedule, prevNextRunAt, tick time.Time) int {
	n := missedRuns(sched, prevNextRunAt, tick)
	if n <= 0 {
		return 0
	}
	return n - 1
}

// missedRuns 统计 (from, to] 区间内的计划触发次数，用于漏跑判定。
func missedRuns(sched cron.Schedule, from, to time.Time) int {
	if sched == nil || !from.Before(to) {
		return 0
	}
	count := 0
	for t := from; ; {
		t = sched.Next(t)
		if t.IsZero() || t.After(to) {
			return count
		}
		count++
		if count > maxCatchUpRuns {
			return count // 已足够判定“多轮漏跑”，提前结束扫描
		}
	}
}

// helper：cronJob 适配 cron.FuncJob
type cronJob func()

func (f cronJob) Run() { f() }
