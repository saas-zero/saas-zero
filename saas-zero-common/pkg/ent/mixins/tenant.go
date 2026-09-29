package mixins

import (
	"context"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

type TenantMixin struct {
	mixin.Schema
	Optional bool
}

func (t TenantMixin) Fields() []ent.Field {
	f := field.Int64("tenant_id").Comment("租户ID | Tenant ID")
	if t.Optional {
		f = f.Optional().Default(0)
	} else {
		f = f.Positive()
	}
	return []ent.Field{f}
}

func (TenantMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id"),
	}
}

func (t TenantMixin) Hooks() []ent.Hook {
	return []ent.Hook{
		t.tenantIdHook(),
	}
}

func (t TenantMixin) tenantIdHook() ent.Hook {
	return func(next ent.Mutator) ent.Mutator {
		return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
			if m.Op() == ent.OpCreate {
				tenantId := GetCurrentTenantId(ctx)
				if tenantId > 0 {
					m.SetField("tenant_id", tenantId)
				}
			}
			return next.Mutate(ctx, m)
		})
	}
}
