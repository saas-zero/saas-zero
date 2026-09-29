package casbin

import (
	"database/sql"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

const modelText = `
[request_definition]
r = sub, dom, obj, act

[policy_definition]
p = sub, dom, obj, act, ept

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.dom == p.dom && keyMatch(r.obj, p.obj) && regexMatch(r.act, p.act)
`

func NewEnforcer(db *sql.DB, tableName string) (*casbin.SyncedEnforcer, error) {
	m, err := model.NewModelFromString(modelText)
	if err != nil {
		return nil, err
	}
	a := NewPgAdapter(db, tableName)
	e, err := casbin.NewSyncedEnforcer(m, a)
	if err != nil {
		return nil, err
	}
	e.EnableAutoSave(true)
	return e, nil
}
