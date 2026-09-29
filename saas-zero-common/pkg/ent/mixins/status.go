package mixins

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// StatusMixin 定义所有表的通用基础字段
// 适用于 PostgreSQL 数据库
type StatusMixin struct {
	ent.Mixin
}

func (StatusMixin) Fields() []ent.Field {
	return []ent.Field{
		//field.Int("status").Default(0).Comment("go特性int默认未0，0为假，1为真"),
		field.Enum("status").
			Values("active", "inactive", "suspended").
			Default("active").
			Comment("状态：active-有效，inactive-无效，suspended-暂停"),
	}
}

// Edges of the StatusMixin.
func (StatusMixin) Edges() []ent.Edge {
	return nil
}

// Indexes of the StatusMixin.
func (StatusMixin) Indexes() []ent.Index {
	return nil
}

// Hooks of the StatusMixin.
func (StatusMixin) Hooks() []ent.Hook {
	return nil
}

// Interceptors of the StatusMixin.
func (StatusMixin) Interceptors() []ent.Interceptor {
	return nil
}

// Policy of the StatusMixin.
func (StatusMixin) Policy() ent.Policy {
	return nil
}

// Annotations of the StatusMixin.
func (StatusMixin) Annotations() []schema.Annotation {
	return nil
}
