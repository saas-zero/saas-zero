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

type SysDict struct {
	ent.Schema
}

func (SysDict) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(128).Comment("名称 | Name"),
		field.String("key").NotEmpty().MaxLen(128).Comment("键 | Key"),
	}
}

func (SysDict) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("sys_dict_datas", SysDictData.Type),
	}
}

func (SysDict) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.TenantMixin{Optional: true},
		mixins.CreatedMixin{},
		mixins.UpdatedMixin{},
		mixins.DeletedMixin{},
		mixins.StatusMixin{},
		mixins.RemarkMixin{},
	}
}

func (SysDict) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Dictionary Table | 字典表"),
		entsql.Annotation{Table: "sys_dicts"},
	}
}

func (SysDict) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "key").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("tenant_id", "status"),
		index.Fields("tenant_id", "name"),
	}
}
