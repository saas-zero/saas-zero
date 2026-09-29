package svc

import (
	"github.com/saas-zero/saas-zero-basedata/ent"
	"github.com/saas-zero/saas-zero-common/pkg/redis"
)

// BumpUsersTokenVersion 递增一批用户的 tokenVersion，踢掉他们的旧会话。
// 任一失败即返回错误：权限/状态已变更但会话未失效，不能对调用方报告成功。
func BumpUsersTokenVersion(store redis.TokenVersionStore, users []*ent.SysUser) error {
	ids := make([]int64, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	return redis.BumpTokenVersions(store, ids...)
}
