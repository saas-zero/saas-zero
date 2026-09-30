// Copyright (c) [2025] Kong All rights reserved.
// Use of this source code is governed by a Apache 2.0 license that can be found in the LICENSE file.
// Author: Kong See：https://github.com/saas-zero/saas-zero or https://gitee.com/saas-zero/saas-zero
// Email: hot_kun@hotmail.com

package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
)

// SysJobLog 定时任务执行日志表 | Scheduled Job Execution Log Table
// 只增不改不软删：由调度器写入，按 created_at 定期物理清理。
type SysJobLog struct {
	ent.Schema
}

func (SysJobLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("job_id").Positive().Comment("任务ID | Job ID"),
		field.String("job_name").Default("").MaxLen(128).Comment("任务名称(快照) | Job Name(snapshot)"),
		field.String("job_group").Default("").MaxLen(64).Comment("任务分组(快照) | Job Group(snapshot)"),
		field.String("handler").Default("").MaxLen(128).Comment("处理器注册码(快照) | Handler Code(snapshot)"),
		field.Enum("trigger_type").Values("cron", "manual", "retry").Default("cron").Comment("触发方式: cron-调度, manual-手动, retry-重试 | Trigger Type"),
		field.String("exec_node").Default("").MaxLen(128).Comment("执行实例标识(hostname:pid，容器主机名可能较长) | Executor Node"),
		field.Enum("status").Values("running", "success", "fail", "timeout", "skipped").Default("running").Comment("执行状态 | Execution Status"),
		field.Int32("attempt").Default(1).NonNegative().Comment("尝试次数(1=首次) | Attempt"),
		field.String("message").Default("").MaxLen(500).Comment("结果摘要 | Result Message"),
		field.String("exception_info").Default("").MaxLen(1000).Comment("异常信息(截断) | Exception Info(truncated)"),
		field.Int64("duration").Default(0).NonNegative().Comment("执行时长(毫秒) | Duration(ms)"),
	}
}

func (SysJobLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.CreatedMixin{},
	}
}

func (SysJobLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("job_log Table | 定时任务执行日志表"),
		entsql.Annotation{Table: "sys_job_logs"}}
}

func (SysJobLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("job_id", "created_at"),
		index.Fields("created_at"),
	}
}
