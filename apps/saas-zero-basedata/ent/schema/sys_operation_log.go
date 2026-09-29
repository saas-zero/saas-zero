package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
)

type SysOperationLog struct {
	ent.Schema
}

func (SysOperationLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("module").Default("").MaxLen(64).Comment("模块 | Module"),
		field.String("operation").Default("").MaxLen(128).Comment("操作 | Operation"),
		field.String("method").Default("").MaxLen(12).Comment("请求方式 | HTTP Method"),
		field.String("path").Default("").MaxLen(255).Comment("请求路径 | Request Path"),
		field.String("params").Optional().Comment("请求参数 | Request Params"),
		field.String("result").Optional().Comment("返回结果 | Response Result"),
		field.String("error_msg").Optional().Comment("失败原因 | Error Message"),
		field.String("status").Default("success").MaxLen(16).Comment("执行状态 success/failure | Execute Status"),
		field.Int64("duration").Default(0).Comment("耗时(ms) | Duration"),
		field.String("ip").Default("").MaxLen(45).Comment("操作IP | Operator IP"),
		field.String("user_agent").Optional().MaxLen(500).Comment("User-Agent | 用户代理"),
		field.Int64("operator_id").Default(0).Comment("操作人ID | Operator ID"),
		field.String("operator_name").Default("").MaxLen(64).Comment("操作人名称 | Operator Name"),
		field.Int64("tenant_id").Default(0).Optional().Comment("租户ID | Tenant ID"),
		field.Time("created_at").Default(time.Now).Immutable().Comment("操作时间 | Operate Time"),
	}
}

func (SysOperationLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

func (SysOperationLog) Edges() []ent.Edge {
	return nil
}

func (SysOperationLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Operation Log Table | 操作日志表"),
		entsql.Annotation{Table: "sys_operation_logs"},
	}
}

func (SysOperationLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("module", "operation"),
		index.Fields("operator_id"),
		index.Fields("tenant_id"),
		index.Fields("path"),
	}
}
