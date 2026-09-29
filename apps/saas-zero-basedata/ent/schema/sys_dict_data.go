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

type SysDictData struct {
	ent.Schema
}

func (SysDictData) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(128).Comment("名称 | Name"),
		field.String("key").NotEmpty().MaxLen(128).Comment("键 | Key"),
		field.String("value").NotEmpty().MaxLen(128).Comment("值 | Value"),
		field.Int64("dict_id").Positive().Comment("字典ID | Dictionary ID"),
	}
}

func (SysDictData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("sys_dict", SysDict.Type).Ref("sys_dict_datas").Unique().Required().Field("dict_id"),
	}
}

func (SysDictData) Mixin() []ent.Mixin {
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

func (SysDictData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Dictionary Data Table | 字典数据表"),
		entsql.Annotation{Table: "sys_dict_datas"},
	}
}

func (SysDictData) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("dict_id"),
		index.Fields("status"),
		index.Fields("tenant_id", "dict_id", "key").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("tenant_id", "dict_id", "status"),
	}
}
