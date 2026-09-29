package types

// Job 任务定义（与 RPC Job 对齐，id 用 string 防前端精度丢失）
type Job struct {
	Id             string `json:"id" form:"id"`
	CreatedAt      int64  `json:"createdAt,omitempty"`
	UpdatedAt      int64  `json:"updatedAt,omitempty"`
	Name           string `json:"name,omitempty"`
	Group          string `json:"group,omitempty"`
	Handler        string `json:"handler,omitempty"`
	Params         string `json:"params,omitempty"`
	CronExpression string `json:"cronExpression" form:"cronExpression"`
	TimeZone       string `json:"timeZone,omitempty"`
	MisfirePolicy  string `json:"misfirePolicy,omitempty"`
	Concurrent     bool   `json:"concurrent" form:"concurrent"`
	Timeout        int32  `json:"timeout,omitempty"`
	MaxRetry       int32  `json:"maxRetry,omitempty"`
	RetryInterval  int32  `json:"retryInterval,omitempty"`
	Status         string `json:"status,omitempty"`
	Remark         string `json:"remark,omitempty"`
	NextRunAt      int64  `json:"nextRunAt,omitempty"`
	LastRunAt      int64  `json:"lastRunAt,omitempty"`
	LastStatus     string `json:"lastStatus,omitempty"`
	LastDuration   int64  `json:"lastDuration,omitempty"`
	LastError      string `json:"lastError,omitempty"`
}

// IdReq 单 ID 请求（detail/start/pause/runOnce 用 query id）
type IdReq struct {
	Id string `json:"id" form:"id"`
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
	Ids []string `json:"ids"`
}

// CleanJobLogReq 清理日志
type CleanJobLogReq struct {
	JobId    string `json:"jobId"`
	KeepDays int32  `json:"keepDays"`
}
