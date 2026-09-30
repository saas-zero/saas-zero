package types

// Job 任务定义（与 RPC Job 对齐，id 用 string 防前端精度丢失）
// 所有字段必须标 optional：go-zero 的 mapping 只以 `optional` 判定可选
// （omitempty 不被识别），未标 optional 的字段在 httpx.Parse 时一律视为必填，
// 缺失即报 `field "xxx" is not set`（HTTP 500）。
type Job struct {
	Id             string `json:"id,optional" form:"id,optional"`
	CreatedAt      int64  `json:"createdAt,omitempty,optional"`
	UpdatedAt      int64  `json:"updatedAt,omitempty,optional"`
	Name           string `json:"name,omitempty,optional"`
	Group          string `json:"group,omitempty,optional"`
	Handler        string `json:"handler,omitempty,optional"`
	Params         string `json:"params,omitempty,optional"`
	CronExpression string `json:"cronExpression,optional" form:"cronExpression,optional"`
	TimeZone       string `json:"timeZone,omitempty,optional"`
	MisfirePolicy  string `json:"misfirePolicy,omitempty,optional"`
	Concurrent     bool   `json:"concurrent,optional" form:"concurrent,optional"`
	Timeout        int32  `json:"timeout,omitempty,optional"`
	MaxRetry       int32  `json:"maxRetry,omitempty,optional"`
	RetryInterval  int32  `json:"retryInterval,omitempty,optional"`
	Status         string `json:"status,omitempty,optional"`
	Remark         string `json:"remark,omitempty,optional"`
	NextRunAt      int64  `json:"nextRunAt,omitempty,optional"`
	LastRunAt      int64  `json:"lastRunAt,omitempty,optional"`
	LastStatus     string `json:"lastStatus,omitempty,optional"`
	LastDuration   int64  `json:"lastDuration,omitempty,optional"`
	LastError      string `json:"lastError,omitempty,optional"`
}

// IdReq 单 ID 请求（detail 走 query id，start/pause/runOnce 走 JSON body id）
// 必填由 RPC 校验（id is required），HTTP 层只负责解析，故标 optional。
type IdReq struct {
	Id string `json:"id,optional" form:"id,optional"`
}

// JobLog 任务执行日志
type JobLog struct {
	Id            string `json:"id"`
	CreatedAt     int64  `json:"createdAt"`
	JobId         string `json:"jobId"`
	JobName       string `json:"jobName"`
	JobGroup      string `json:"jobGroup"`
	Handler       string `json:"handler"`
	TriggerType   string `json:"triggerType"`
	ExecNode      string `json:"execNode"`
	Status        string `json:"status"`
	Attempt       int32  `json:"attempt"`
	Message       string `json:"message"`
	ExceptionInfo string `json:"exceptionInfo"`
	Duration      int64  `json:"duration"`
}

// BaseResp 统一响应（与 basedata API 一致）
type BaseResp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// JobPageReq 任务列表请求
type JobPageReq struct {
	Name     string `form:"name,optional"`
	Group    string `form:"group,optional"`
	Handler  string `form:"handler,optional"`
	Status   string `form:"status,optional"`
	Page     int32  `form:"page,optional,default=1"`
	PageSize int32  `form:"pageSize,optional,default=20"`
}

// JobPageResp 任务列表响应
type JobPageResp struct {
	List  []*Job `json:"list"`
	Total int64  `json:"total"`
}

// JobLogPageReq 任务日志列表请求
type JobLogPageReq struct {
	JobId     string `form:"jobId,optional"`
	JobName   string `form:"jobName,optional"`
	Status    string `form:"status,optional"`
	BeginTime int64  `form:"beginTime,optional"`
	EndTime   int64  `form:"endTime,optional"`
	Page      int32  `form:"page,optional,default=1"`
	PageSize  int32  `form:"pageSize,optional,default=20"`
}

// JobLogPageResp 任务日志列表响应
type JobLogPageResp struct {
	List  []*JobLog `json:"list"`
	Total int64     `json:"total"`
}

// IdsReq 批量删除
type IdsReq struct {
	Ids []string `json:"ids,optional"`
}

// CleanJobLogReq 清理日志
type CleanJobLogReq struct {
	JobId    string `json:"jobId,optional"`
	KeepDays int32  `json:"keepDays,optional"`
}
