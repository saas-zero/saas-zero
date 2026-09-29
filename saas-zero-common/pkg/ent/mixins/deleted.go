package mixins

import (
	"context"
	"strconv"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type DeletedMixin struct {
	ent.Mixin
}

func (DeletedMixin) Fields() []ent.Field {
	return []ent.Field{
		// 时间字段
		field.Time("deleted_at").
			Optional().
			Comment("deleted Time | 删除时间"),

		// 操作人 ID
		field.Int64("deleted_id").
			Optional().
			Comment("deleted ID | 删除人ID，逻辑删除用"),

		// 操作人名称
		field.String("deleted_by").
			Optional().
			MaxLen(64).
			Comment("deleted Name | 删除人名称"),
	}
}

// Edges of the DeletedMixin.
func (DeletedMixin) Edges() []ent.Edge {
	return nil
}

// Indexes of the DeletedMixin.
func (DeletedMixin) Indexes() []ent.Index {
	return nil
}

// Hooks of the DeletedMixin.
func (DeletedMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				if !m.Op().Is(ent.OpUpdate) && !m.Op().Is(ent.OpUpdateOne) {
					return next.Mutate(ctx, m)
				}
				if _, exists := m.Field("deleted_at"); !exists {
					return next.Mutate(ctx, m)
				}
				now := time.Now()
				m.SetField("deleted_at", now)
				if uid := GetCurrentUserId(ctx); uid > 0 {
					m.SetField("deleted_id", uid)
				}
				uname := GetCurrentUserName(ctx)
				if uname == "" {
					if uid := GetCurrentUserId(ctx); uid > 0 {
						uname = strconv.FormatInt(uid, 10)
					}
				}
				if uname != "" {
					m.SetField("deleted_by", uname)
				}
				return next.Mutate(ctx, m)
			})
		},
	}
}

// Interceptors of the DeletedMixin.
func (DeletedMixin) Interceptors() []ent.Interceptor {
	return nil
}

// Policy of the DeletedMixin.
func (DeletedMixin) Policy() ent.Policy {
	return nil
}

// Annotations of the DeletedMixin.
func (DeletedMixin) Annotations() []schema.Annotation {
	return nil
}
