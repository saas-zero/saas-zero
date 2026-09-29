package mixins

import (
	"context"
	"strconv"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// CreatedMixin 定义所有表的通用基础字段
// 适用于 PostgreSQL 数据库

// CreatedMixin 定义所有表的通用基础字段
// 适用于 PostgreSQL 数据库
type CreatedMixin struct {
	ent.Mixin
}

func (CreatedMixin) Fields() []ent.Field {
	return []ent.Field{
		// 时间字段
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("Create Time | 创建时间"),

		// 操作人 ID
		field.Int64("created_id").
			Positive().
			Immutable().
			Comment("Creator ID | 创建人ID"),

		// 操作人名称
		field.String("created_by").
			NotEmpty().
			MaxLen(64).
			Immutable().
			Comment("Creator Name | 创建人名称"),
	}
}

// Edges of the CreatedMixin.
func (CreatedMixin) Edges() []ent.Edge {
	return nil
}

// Indexes of the CreatedMixin.
func (CreatedMixin) Indexes() []ent.Index {
	return nil
}

// Hooks of the CreatedMixin.
func (CreatedMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		func(next ent.Mutator) ent.Mutator {
			return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
				if !m.Op().Is(ent.OpCreate) {
					return next.Mutate(ctx, m)
				}
				now := time.Now()
				m.SetField("created_at", now)
				if uid := GetCurrentUserId(ctx); uid > 0 {
					m.SetField("created_id", uid)
				}
				uname := GetCurrentUserName(ctx)
				if uname == "" {
					if uid := GetCurrentUserId(ctx); uid > 0 {
						uname = strconv.FormatInt(uid, 10)
					}
				}
				if uname != "" {
					m.SetField("created_by", uname)
				}
				return next.Mutate(ctx, m)
			})
		},
	}
}

// Interceptors of the CreatedMixin.
func (CreatedMixin) Interceptors() []ent.Interceptor {
	return nil
}

// Policy of the CreatedMixin.
func (CreatedMixin) Policy() ent.Policy {
	return nil
}

// Annotations of the CreatedMixin.
func (CreatedMixin) Annotations() []schema.Annotation {
	return nil
}
