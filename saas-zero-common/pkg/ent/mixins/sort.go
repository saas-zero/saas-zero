package mixins

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// SortMixin 定义所有表的通用基础字段
// 适用于 PostgreSQL 数据库
type SortMixin struct {
	ent.Mixin
}

func (SortMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Uint32("sort").
			Default(1).Positive().
			Comment("Sort Number | 排序编号"),
	}
}

// Edges of the SortMixin.
func (SortMixin) Edges() []ent.Edge {
	return nil
}

// Indexes of the SortMixin.
func (SortMixin) Indexes() []ent.Index {
	return nil
}

// Hooks of the SortMixin.
func (SortMixin) Hooks() []ent.Hook {
	return nil
}

// Interceptors of the SortMixin.
func (SortMixin) Interceptors() []ent.Interceptor {
	return nil
}

// Policy of the SortMixin.
func (SortMixin) Policy() ent.Policy {
	return nil
}

// Annotations of the SortMixin.
func (SortMixin) Annotations() []schema.Annotation {
	return nil
}
