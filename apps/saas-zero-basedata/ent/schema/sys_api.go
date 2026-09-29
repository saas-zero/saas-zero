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

type SysApi struct {
	ent.Schema
}

func (SysApi) Fields() []ent.Field {
	return []ent.Field{
		field.String("api_name").Default("").MaxLen(128).Comment("API名称 | API Name"),
		field.Enum("api_type").Values("group", "api").Default("group").Comment("类型：group-分组 api-接口 | API Type"),
		field.String("api_path").Default("").MaxLen(128).Comment("API路径 | API Path"),
		field.Enum("api_method").Values("get", "post", "put", "delete").Optional().Comment("请求方法 | HTTP Method"),
	}
}

func (SysApi) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("packages", SysPackage.Type).Ref("apis"),
	}
}

func (SysApi) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.CreatedMixin{},
		mixins.UpdatedMixin{},
		mixins.DeletedMixin{},
		mixins.StatusMixin{},
		mixins.RemarkMixin{},
	}
}

func (SysApi) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("API Table | 接口表"),
		entsql.Annotation{Table: "sys_apis"},
	}
}

func (SysApi) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("api_path", "api_method").Unique().StorageKey("idx_api_path_method_unique").Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("api_type"),
		index.Fields("status"),
	}
}
