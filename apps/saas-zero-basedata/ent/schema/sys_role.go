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

type SysRole struct {
	ent.Schema
}

func (SysRole) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(128).Comment("名称 | Name"),
		field.String("code").NotEmpty().MaxLen(128).Comment("编码 | Code"),
		field.Bool("is_system").Default(false).Comment("系统内置角色 | System Role（系统角色不可删除/修改）"),
	}
}

func (SysRole) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("menus", SysMenu.Type),
		edge.From("users", SysUser.Type).Ref("roles"),
	}
}

func (SysRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.TenantMixin{},
		mixins.CreatedMixin{},
		mixins.UpdatedMixin{},
		mixins.DeletedMixin{},
		mixins.StatusMixin{},
		mixins.SortMixin{},
		mixins.RemarkMixin{},
	}
}

func (SysRole) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Role Table | 角色表"),
		entsql.Annotation{Table: "sys_roles"},
	}
}

func (SysRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "code").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("tenant_id", "status"),
		index.Fields("sort"),
	}
}
