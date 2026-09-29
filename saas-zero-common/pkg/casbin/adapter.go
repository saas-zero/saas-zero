package casbin

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
)

type PgAdapter struct {
	db        *sql.DB
	tableName string
}

func NewPgAdapter(db *sql.DB, tableName string) *PgAdapter {
	if tableName == "" {
		tableName = "casbin_rule"
	}
	a := &PgAdapter{db: db, tableName: tableName}
	a.createTable()
	return a
}

func (a *PgAdapter) createTable() {
	q := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id SERIAL PRIMARY KEY,
		ptype VARCHAR(100) NOT NULL DEFAULT '',
		v0 VARCHAR(100) NOT NULL DEFAULT '',
		v1 VARCHAR(100) NOT NULL DEFAULT '',
		v2 VARCHAR(100) NOT NULL DEFAULT '',
		v3 VARCHAR(100) NOT NULL DEFAULT '',
		v4 VARCHAR(100) NOT NULL DEFAULT '',
		v5 VARCHAR(100) NOT NULL DEFAULT ''
	)`, a.tableName)
	a.db.Exec(q)
}

func (a *PgAdapter) LoadPolicy(m model.Model) error {
	q := fmt.Sprintf("SELECT ptype, v0, v1, v2, v3, v4, v5 FROM %s", a.tableName)
	rows, err := a.db.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var ptype, v0, v1, v2, v3, v4, v5 string
		if err := rows.Scan(&ptype, &v0, &v1, &v2, &v3, &v4, &v5); err != nil {
			return err
		}
		line := fmt.Sprintf("%s, %s, %s, %s, %s, %s", ptype, v0, v1, v2, v3, v4)
		if err := persist.LoadPolicyLine(line, m); err != nil {
			return err
		}
	}
	return nil
}

func (a *PgAdapter) SavePolicy(m model.Model) error {
	return nil
}

func (a *PgAdapter) AddPolicy(sec, ptype string, rule []string) error {
	cols := make([]string, 6)
	cols[0] = ptype
	for i := 0; i < 5 && i < len(rule); i++ {
		cols[i+1] = rule[i]
	}
	q := fmt.Sprintf("INSERT INTO %s (ptype, v0, v1, v2, v3, v4) VALUES ($1,$2,$3,$4,$5,$6)", a.tableName)
	args := []interface{}{cols[0], cols[1], cols[2], cols[3], cols[4], cols[5]}
	_, err := a.db.Exec(q, args...)
	return err
}

func (a *PgAdapter) RemovePolicy(sec, ptype string, rule []string) error {
	cols := make([]string, 6)
	cols[0] = ptype
	for i := 0; i < 5 && i < len(rule); i++ {
		cols[i+1] = rule[i]
	}
	q := fmt.Sprintf("DELETE FROM %s WHERE ptype=$1 AND v0=$2 AND v1=$3 AND v2=$4 AND v3=$5 AND v4=$6", a.tableName)
	args := []interface{}{cols[0], cols[1], cols[2], cols[3], cols[4], cols[5]}
	_, err := a.db.Exec(q, args...)
	return err
}

func (a *PgAdapter) RemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error {
	var wheres []string
	args := []interface{}{ptype}
	for i, v := range fieldValues {
		if v == "" {
			continue
		}
		col := fmt.Sprintf("v%d", fieldIndex+i)
		wheres = append(wheres, fmt.Sprintf("%s=$%d", col, len(args)+1))
		args = append(args, v)
	}
	if len(wheres) == 0 {
		return nil
	}
	q := fmt.Sprintf("DELETE FROM %s WHERE ptype=$1 AND %s", a.tableName, strings.Join(wheres, " AND "))
	_, err := a.db.Exec(q, args...)
	return err
}
