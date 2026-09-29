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

type SysPackage struct {
	ent.Schema
}

func (SysPackage) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(128).Comment("套餐名称 | Package Name"),
		field.String("code").NotEmpty().MaxLen(64).Comment("套餐编码 | Package Code"),
	}
}

func (SysPackage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("menus", SysMenu.Type),
		edge.To("apis", SysApi.Type),
		edge.To("tenants", SysTenant.Type),
	}
}

func (SysPackage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.CreatedMixin{},
		mixins.UpdatedMixin{},
		mixins.DeletedMixin{},
		mixins.StatusMixin{},
		mixins.SortMixin{},
		mixins.RemarkMixin{},
	}
}

func (SysPackage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Package Table | 套餐表"),
		entsql.Annotation{Table: "sys_packages"},
	}
}

func (SysPackage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("sort"),
		// 软删除：唯一约束只对未删除行生效
		index.Fields("code").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
	}
}
