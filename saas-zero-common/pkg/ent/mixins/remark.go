package mixins

import (
	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RemarkMixin 定义所有表的通用基础字段
// 适用于 PostgreSQL 数据库
type RemarkMixin struct {
	ent.Mixin
}

func (RemarkMixin) Fields() []ent.Field {
	return []ent.Field{
		field.String("remark").MaxLen(255).Optional(),
	}
}

// Edges of the RemarkMixin.
func (RemarkMixin) Edges() []ent.Edge {
	return nil
}

// Indexes of the RemarkMixin.
func (RemarkMixin) Indexes() []ent.Index {
	return nil
}

// Hooks of the RemarkMixin.
func (RemarkMixin) Hooks() []ent.Hook {
	return nil
}

// Interceptors of the RemarkMixin.
func (RemarkMixin) Interceptors() []ent.Interceptor {
	return nil
}

// Policy of the RemarkMixin.
func (RemarkMixin) Policy() ent.Policy {
	return nil
}

// Annotations of the RemarkMixin.
func (RemarkMixin) Annotations() []schema.Annotation {
	return nil
}
