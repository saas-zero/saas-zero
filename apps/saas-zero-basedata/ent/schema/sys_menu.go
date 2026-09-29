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

type SysMenu struct {
	ent.Schema
}

func (SysMenu) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("menu_type").
			Values("directory", "menu", "button").
			Default("directory").
			Comment("类型：directory-目录 menu-菜单 button-按钮 | Menu Type"),
		field.String("name").NotEmpty().MaxLen(128).Comment("名称 | Name"),
		field.Int64("parent_id").Default(0).Optional().Comment("父级ID | Parent ID"),
		field.String("component").Optional().Default("").Comment("组件路径 | Component Path"),
		field.String("path").Optional().Default("").Comment("路由路径 | Route Path"),
		field.String("icon").Default("").Comment("图标 | Icon"),
		field.Bool("is_redirect").Default(false).Comment("是否重定向 | Is Redirect"),
		field.String("redirect").Optional().Default("").Comment("重定向路径 | Redirect Path"),
		field.Bool("hidden").Default(false).Comment("是否隐藏 | Hidden"),
	}
}

func (SysMenu) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("roles", SysRole.Type).Ref("menus"),
		edge.From("packages", SysPackage.Type).Ref("menus"),
	}
}

func (SysMenu) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.CreatedMixin{},
		mixins.UpdatedMixin{},
		mixins.DeletedMixin{},
		mixins.StatusMixin{},
		mixins.RemarkMixin{},
		mixins.SortMixin{},
	}
}

func (SysMenu) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("Menu Table | 菜单表"),
		entsql.Annotation{Table: "sys_menus"},
	}
}

func (SysMenu) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parent_id", "sort"),
		index.Fields("path"),
		index.Fields("name"),
		index.Fields("parent_id"),
		index.Fields("menu_type", "status"),
	}
}
