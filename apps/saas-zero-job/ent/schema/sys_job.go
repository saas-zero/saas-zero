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

// SysJob 定时任务表 | Scheduled Job Table
type SysJob struct {
	ent.Schema
}

func (SysJob) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(128).Comment("任务名称 | Job Name"),
		field.String("group").Default("default").MaxLen(64).Comment("任务分组(字典 job_group) | Job Group"),
		field.String("handler").NotEmpty().MaxLen(128).Comment("处理器注册码(代码注册表, 如 report.daily) | Handler Code"),
		field.Text("params").Default("{}").Comment("处理器参数(JSON) | Handler Params(JSON)"),
		field.String("cron_expression").NotEmpty().MaxLen(255).Comment("cron表达式(兼容5/6位, 秒可选) | Cron Expression(5/6 fields, seconds optional)"),
		field.String("time_zone").Default("Asia/Shanghai").MaxLen(64).Comment("表达式时区 | Time Zone"),
		field.Enum("misfire_policy").Values("skip", "fire_once", "fire_all").Default("fire_once").Comment("错过执行策略: skip-放弃, fire_once-补跑一次, fire_all-全部补跑 | Misfire Policy"),
		field.Bool("concurrent").Default(false).Comment("是否允许并发执行(否-同一时刻只跑一次) | Allow Concurrent"),
		field.Int32("timeout").Default(30).NonNegative().Comment("单次执行超时(秒) | Timeout(seconds)"),
		field.Int32("max_retry").Default(0).NonNegative().Comment("失败重试次数 | Max Retry"),
		field.Int32("retry_interval").Default(10).NonNegative().Comment("重试间隔(秒) | Retry Interval(seconds)"),
		field.Time("next_run_at").Optional().Comment("下次触发时间(调度器回写) | Next Run At"),
		field.Time("last_run_at").Optional().Comment("上次触发时间 | Last Run At"),
		field.String("last_status").Optional().MaxLen(16).Comment("上次执行结果: success/fail/timeout/skipped/none | Last Status"),
		field.Int64("last_duration").Optional().Comment("上次耗时(毫秒) | Last Duration(ms)"),
		field.String("last_error").Optional().MaxLen(500).Comment("上次错误摘要 | Last Error"),
	}
}

func (SysJob) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.CreatedMixin{},
		mixins.UpdatedMixin{},
		mixins.DeletedMixin{},
		mixins.StatusMixin{},
		mixins.RemarkMixin{},
	}
}

func (SysJob) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("job Table | 定时任务表"),
		entsql.Annotation{Table: "sys_jobs"}}
}

func (SysJob) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status", "next_run_at").StorageKey("idx_jobs_status_next_run"),
		index.Fields("group"),
	}
}
