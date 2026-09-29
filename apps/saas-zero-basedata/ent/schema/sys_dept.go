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

type SysDept struct {
	ent.Schema
}

func (SysDept) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(128).Comment("名称 | Name"),
		field.Int64("leader_id").Optional().Comment("负责人ID | Leader ID"),
		field.String("mobile").Default("").MaxLen(20).Comment("部门电话 | Department Phone"),
		field.String("email").Default("").MaxLen(64).Comment("部门邮箱 | Department Email"),
		field.Int64("parent_id").Default(0).Optional().Comment("父级ID | Parent ID"),
		field.String("parent_name").Default("").MaxLen(128).Comment("父级名称 | Parent Name"),
	}
}

func (SysDept) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("sys_tenant", SysTenant.Type).Ref("sys_depts").Unique().Required().Field("tenant_id"),
		edge.To("leader", SysUser.Type).Field("leader_id").Unique(),
		edge.To("sys_users", SysUser.Type),
	}
}

func (SysDept) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.TenantMixin{},
		mixins.CreatedMixin{},
		mixins.UpdatedMixin{},
		mixins.DeletedMixin{},
		mixins.StatusMixin{},
		mixins.SortMixin{},
	}
}

func (SysDept) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Department Table | 部门表"),
		entsql.Annotation{Table: "sys_depts"},
	}
}

func (SysDept) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "parent_id"),
		index.Fields("tenant_id", "name"),
		index.Fields("tenant_id", "status"),
	}
}
