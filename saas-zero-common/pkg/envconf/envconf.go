// Package envconf 提供环境变量覆盖配置项的工具方法。
//
// 设计目的：YAML 明文配置便于本地调试，环境变量用于生产环境覆盖敏感配置
// （JWT 密钥、Redis/PostgreSQL 连接等），两者互不影响。
package envconf

import "os"

// String 返回环境变量 key 的值；未设置或为空时回退到 fallback（来自明文配置文件）。
func String(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
