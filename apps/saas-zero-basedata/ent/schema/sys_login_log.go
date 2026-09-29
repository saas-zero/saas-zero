package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
)

type SysLoginLog struct {
	ent.Schema
}

func (SysLoginLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("user_id").Positive().Comment("用户ID | User ID"),
		field.String("username").NotEmpty().MaxLen(128).Comment("用户名 | Username"),
		field.String("ip").Default("").MaxLen(45).Comment("登录IP | Login IP"),
		field.Enum("status").Values("success", "fail").Default("success").Comment("状态：success-成功 fail-失败 | Status"),
		field.String("message").Default("").MaxLen(255).Comment("消息 | Message"),
		field.Time("login_time").Comment("登录时间 | Login Time"),
		field.Int64("tenant_id").Default(0).Optional().Comment("租户ID | Tenant ID"),
	}
}

func (SysLoginLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

func (SysLoginLog) Edges() []ent.Edge {
	return nil
}

func (SysLoginLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Login Log Table | 登录日志表"),
		entsql.Annotation{Table: "sys_login_logs"},
	}
}

func (SysLoginLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("username"),
		index.Fields("login_time"),
		index.Fields("tenant_id"),
	}
}
