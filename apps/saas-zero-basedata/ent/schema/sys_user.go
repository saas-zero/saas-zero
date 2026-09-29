package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/saas-zero/saas-zero-common/pkg/ent/mixins"
)

type SysUser struct {
	ent.Schema
}

func (SysUser) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").Immutable().NotEmpty().MaxLen(128).Comment("用户名 | Username"),
		field.String("password").NotEmpty().MaxLen(255).Comment("密码 | Password"),
		field.String("nickname").Default("").MaxLen(128).Comment("昵称 | Nickname"),
		field.String("mobile").Default("").MaxLen(20).Comment("手机号码 | Mobile"),
		field.String("email").Default("").MaxLen(64).Comment("邮箱 | Email"),
		field.Int64("dept_id").Optional().Comment("部门ID | Department ID"),
		field.String("login_ip").Default("").MaxLen(64).Comment("最后登录IP | Last Login IP"),
		field.Time("login_at").Optional().Comment("最后登录时间 | Last Login At"),
		field.Int32("login_error_count").Default(0).Comment("连续登录错误次数 | Login Error Count"),
		field.Time("lockout_until").Optional().Comment("锁定截止时间 | Lockout Until"),
		field.String("position").Default("").MaxLen(64).Comment("岗位 | Position"),
	}
}

func (SysUser) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.TenantMixin{},
		mixins.CreatedMixin{},
		mixins.UpdatedMixin{},
		mixins.DeletedMixin{},
		mixins.StatusMixin{},
		mixins.RemarkMixin{},
	}
}

func (SysUser) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("sys_tenant", SysTenant.Type).Ref("sys_users").Unique().Required().Field("tenant_id"),
		edge.From("sys_dept", SysDept.Type).Ref("sys_users").Unique().Field("dept_id"),
		edge.To("roles", SysRole.Type),
	}
}

func (SysUser) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("User Table | 用户表"),
		entsql.Annotation{Table: "sys_users"},
	}
}

func (SysUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "username").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("tenant_id", "mobile"),
		index.Fields("tenant_id", "email"),
		index.Fields("dept_id"),
		index.Fields("status"),
		index.Fields("created_at"),
		index.Fields("tenant_id", "status"),
		index.Fields("tenant_id", "dept_id"),
	}
}
