package mixins

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/saas-zero/saas-zero-common/pkg/snowflake"
)

// BaseMixin 定义所有表的通用基础字段
// 适用于 PostgreSQL 数据库
type BaseMixin struct {
	ent.Mixin
}

func (BaseMixin) Fields() []ent.Field {
	return []ent.Field{
		// 主键ID：雪花ID，可手动赋值
		field.Int64("id").
			Unique().
			Positive().
			Immutable().
			Comment("Primary Key | 主键ID，可自定义雪花ID"),
	}
}

// Edges of the BaseMixin.
func (BaseMixin) Edges() []ent.Edge {
	return nil
}

// Indexes of the BaseMixin.
func (BaseMixin) Indexes() []ent.Index {
	return nil
}

type idSettable interface {
	SetID(int64)
}

// Hooks of the BaseMixin.
func (BaseMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				if !m.Op().Is(ent.OpCreate) {
					return next.Mutate(ctx, m)
				}
				if s, ok := m.(idSettable); ok {
					s.SetID(snowflake.NextID())
				}
				return next.Mutate(ctx, m)
			})
		},
	}
}

// Interceptors of the BaseMixin.
func (BaseMixin) Interceptors() []ent.Interceptor {
	return nil
}

// Policy of the BaseMixin.
func (BaseMixin) Policy() ent.Policy {
	return nil
}

// Annotations of the BaseMixin.
func (BaseMixin) Annotations() []schema.Annotation {
	return nil
}
