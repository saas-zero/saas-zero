package redis

import "fmt"

// TokenVersionStore 抽象 token_version 递增操作。
// 生产由 *Client 实现；测试可注入失败桩，验证"会话失效失败不得当成成功"。
type TokenVersionStore interface {
	Incr(key string) (int64, error)
}

// TokenVersionKey 返回用户 tokenVersion 的 Redis key。
func TokenVersionKey(userId int64) string {
	return fmt.Sprintf("token_version:%d", userId)
}

// BumpTokenVersion 递增用户 tokenVersion，使其已签发的所有 token 立即失效。
// 密码变更、权限变更、账号禁用/删除后必须调用；失败时返回错误，
// 调用方不得把整体操作报告为成功，否则会出现"数据已变更但旧会话仍有效"的越权窗口。
func BumpTokenVersion(store TokenVersionStore, userId int64) error {
	if store == nil {
		return fmt.Errorf("%w: token_version of user %d", ErrNotInitialized, userId)
	}
	if _, err := store.Incr(TokenVersionKey(userId)); err != nil {
		return fmt.Errorf("bump token_version of user %d: %w", userId, err)
	}
	return nil
}

// BumpTokenVersions 批量递增，任一失败立即返回。
// 已成功递增的用户保持失效状态（偏向安全），调用方仍需把错误上抛。
func BumpTokenVersions(store TokenVersionStore, userIds ...int64) error {
	for _, userId := range userIds {
		if err := BumpTokenVersion(store, userId); err != nil {
			return err
		}
	}
	return nil
}
